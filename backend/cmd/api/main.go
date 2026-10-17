// Command api is the backend of the personal site: it imports content into
// PostgreSQL and serves the read-only JSON API (see specs/backend/spec.md).
package main

import (
	"context"
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
	"github.com/leonid6372/my-site/backend/internal/content"
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
	{"serve", "run migrations and start the HTTP server", runServe},
	{"seed", "validate content/ and replace all content in the database", runSeed},
	{"validate", "validate content/ only (no database access)", runValidate},
	{"migrate", "run database migrations only", runMigrate},
	{"start", "seed, then serve (container default command)", runStart},
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

func runValidate(_ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := loadContent(cfg.ContentDir, os.Stderr); err != nil {
		return err
	}
	fmt.Println("content OK")
	return nil
}

// loadContent loads and validates the content files, printing every
// validation error to w as "file:path: message".
func loadContent(dir string, w io.Writer) (content.All, error) {
	all, err := content.Load(dir)
	if err != nil {
		return content.All{}, err
	}
	if errs := content.Validate(all); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(w, e)
		}
		return content.All{}, fmt.Errorf("content has %d validation error(s)", len(errs))
	}
	return all, nil
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

func runSeed(_ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	setupLogger(cfg.LogLevel)
	ctx := context.Background()
	s, err := openMigrated(ctx, cfg)
	if err != nil {
		return err
	}
	defer s.Close()
	return seed(ctx, s, cfg.ContentDir)
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

// seed loads and validates content/ and replaces all database content.
// Nothing is written unless the content is valid.
func seed(ctx context.Context, s *store.Store, dir string) error {
	all, err := loadContent(dir, os.Stderr)
	if err != nil {
		return err
	}
	if err := s.ReplaceAll(ctx, all); err != nil {
		return err
	}
	slog.Info("content seeded", "links", len(all.Main.Links), "jobs", len(all.CV.Experience),
		"places", len(all.Travel.Places))
	return nil
}

func runServe(_ []string) error { return serveCommand(false) }
func runStart(_ []string) error { return serveCommand(true) }

// serveCommand migrates (and seeds, for start) and then serves until
// SIGINT/SIGTERM. The server never starts if the seed fails.
func serveCommand(withSeed bool) error {
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
	if withSeed {
		if err := seed(ctx, s, cfg.ContentDir); err != nil {
			return err
		}
	}

	handler := httpapi.New(httpapi.Options{
		Cache:       service.NewCache(service.New(s)),
		DB:          s,
		Log:         log,
		CORSOrigins: cfg.CORSOrigins,
	})
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	log.Info("listening", "addr", ln.Addr().String(), "cors_origins", cfg.CORSOrigins)
	return serveHTTP(ctx, ln, handler, log)
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

// serveHTTP serves on ln until ctx is cancelled, then shuts down gracefully,
// waiting up to 10 s for in-flight requests.
func serveHTTP(ctx context.Context, ln net.Listener, h http.Handler, log *slog.Logger) error {
	srv := &http.Server{
		Handler:           h,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

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
// ":8080" or "0.0.0.0:8080" → "http://127.0.0.1:8080/api/v1/health".
func healthURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host, port = "", "8080"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/api/v1/health"
}

// healthcheck returns nil if url answers 200 within 2 s.
func healthcheck(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
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
