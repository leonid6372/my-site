// Package config reads the service configuration from environment variables
// (spec §4).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config is the runtime configuration of the api binary.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	ContentDir  string
	CORSOrigins []string
	LogLevel    slog.Level
	// TLSCertFile and TLSKeyFile are PEM files for HTTPS. Both empty means
	// the server uses a self-signed certificate generated at startup.
	TLSCertFile string
	TLSKeyFile  string
}

// ErrMissingDatabaseURL is returned by RequireDatabase when DATABASE_URL is empty.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")

// Load reads the configuration using os.Getenv.
func Load() (Config, error) {
	return LoadFrom(os.Getenv)
}

// LoadFrom reads the configuration with the given lookup function, so tests
// don't have to touch the process environment. DATABASE_URL is not checked
// here because commands like validate don't need it; see RequireDatabase.
func LoadFrom(getenv func(string) string) (Config, error) {
	cfg := Config{
		HTTPAddr:    valueOr(getenv("HTTP_ADDR"), ":8080"),
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
		ContentDir:  valueOr(getenv("CONTENT_DIR"), "/app/content"),
		CORSOrigins: splitList(getenv("CORS_ORIGINS")),
		TLSCertFile: strings.TrimSpace(getenv("TLS_CERT_FILE")),
		TLSKeyFile:  strings.TrimSpace(getenv("TLS_KEY_FILE")),
	}
	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return Config{}, errors.New("TLS_CERT_FILE and TLS_KEY_FILE must be set together")
	}

	level := strings.ToLower(valueOr(getenv("LOG_LEVEL"), "info"))
	switch level {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		return Config{}, fmt.Errorf("LOG_LEVEL %q must be one of: debug, info, warn, error", level)
	}

	return cfg, nil
}

// RequireDatabase returns an error if the database URL is not configured.
func (c Config) RequireDatabase() error {
	if c.DatabaseURL == "" {
		return ErrMissingDatabaseURL
	}
	return nil
}

func valueOr(v, def string) string {
	if v = strings.TrimSpace(v); v != "" {
		return v
	}
	return def
}

func splitList(v string) []string {
	var out []string
	for part := range strings.SplitSeq(v, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
