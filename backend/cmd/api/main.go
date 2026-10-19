// Command api is the backend of the personal site: it serves the read-only
// JSON API from PostgreSQL (see specs/backend/spec.md). The database is the
// only source of data; the API never writes content.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonid6372/my-site/backend/internal/config"
	"github.com/leonid6372/my-site/backend/internal/httpapi"
	"github.com/leonid6372/my-site/backend/internal/service"
	"github.com/leonid6372/my-site/backend/internal/store"
)

type command struct {
	name  string
	usage string
	run   func(args []string) error
}

var commands = []command{
	{"serve", "run migrations and start the HTTPS server (container default command)", runServe},
	{"migrate", "run database migrations only", runMigrate},
	{"healthcheck", "check the local /api/v1/health endpoint (exit 0 if healthy)", runHealthcheck},
}

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run dispatches to a subcommand and returns the process exit code:
// 0 on success, 1 on command failure, 2 on usage errors.
func run(args []string, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	for _, c := range commands {
		if c.name == args[0] {
			if err := c.run(args[1:]); err != nil {
				fmt.Fprintf(stderr, "api %s: %v\n", c.name, err)
				return 1
			}
			return 0
		}
	}
	fmt.Fprintf(stderr, "api: unknown command %q\n\n", args[0])
	printUsage(stderr)
	return 2
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: api <command>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	for _, c := range commands {
		fmt.Fprintf(w, "  %-12s %s\n", c.name, c.usage)
	}
}

func runMigrate(_ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	s, err := openMigrated(ctx, cfg)
	if err != nil {
		return err
	}
	s.Close()
	fmt.Println("migrations OK")
	return nil
}

// openMigrated connects to the database and applies pending migrations.
func openMigrated(ctx context.Context, cfg config.Config) (*store.Store, error) {
	if err := cfg.RequireDatabase(); err != nil {
		return nil, err
	}
	s, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := s.Migrate(ctx); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

// runServe migrates the schema and serves until SIGINT/SIGTERM. It does not
// add or change any content.
func runServe(_ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := setupLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s, err := openMigrated(ctx, cfg)
	if err != nil {
		return err
	}
	defer s.Close()

	handler := httpapi.New(httpapi.Options{
		Service:     service.New(s),
		DB:          s,
		Log:         log,
		CORSOrigins: cfg.CORSOrigins,
	})
	tlsCfg, err := tlsConfig(cfg, log)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	log.Info("listening (https)", "addr", ln.Addr().String(), "cors_origins", cfg.CORSOrigins)
	return serveHTTP(ctx, ln, handler, tlsCfg, log)
}

// setupLogger installs a JSON slog logger on stdout as the default.
func setupLogger(level slog.Level) *slog.Logger {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)
	return log
}

// Server timeouts (spec §9).
const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
)

// serveHTTP serves HTTPS on ln until ctx is cancelled, then shuts down
// gracefully, waiting up to 10 s for in-flight requests.
func serveHTTP(ctx context.Context, ln net.Listener, h http.Handler, tlsCfg *tls.Config, log *slog.Logger) error {
	srv := &http.Server{
		Handler:           h,
		TLSConfig:         tlsCfg,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ServeTLS(ln, "", "") }() // certificates come from TLSConfig

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("stopped")
	return nil
}

func runHealthcheck(_ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return healthcheck(healthURL(cfg.HTTPAddr))
}

// healthURL turns a listen address into the local health URL, e.g.
// ":8080" or "0.0.0.0:8080" → "https://127.0.0.1:8080/api/v1/health".
func healthURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host, port = "", "8080"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "https://" + net.JoinHostPort(host, port) + "/api/v1/health"
}

// healthcheck returns nil if url answers 200 within 2 s. It only calls the
// local server, whose certificate may be self-signed or issued for a public
// name, so the certificate is not verified.
func healthcheck(url string) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // loopback health probe
		},
	}
	resp, err := client.Get(url) //nolint:gosec // local URL built from our own config
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy: %s", resp.Status)
	}
	return nil
}
