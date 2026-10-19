package config

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := LoadFrom(env(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Config{HTTPAddr: ":8080", LogLevel: slog.LevelInfo}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadValues(t *testing.T) {
	cfg, err := LoadFrom(env(map[string]string{
		"HTTP_ADDR":     "127.0.0.1:9000",
		"DATABASE_URL":  " postgres://u:p@db/site ",
		"CORS_ORIGINS":  " http://localhost:5173 , ,http://example.test",
		"LOG_LEVEL":     "DEBUG",
		"TLS_CERT_FILE": "/certs/cert.pem",
		"TLS_KEY_FILE":  "/certs/key.pem",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Config{
		HTTPAddr:    "127.0.0.1:9000",
		DatabaseURL: "postgres://u:p@db/site",
		CORSOrigins: []string{"http://localhost:5173", "http://example.test"},
		LogLevel:    slog.LevelDebug,
		TLSCertFile: "/certs/cert.pem",
		TLSKeyFile:  "/certs/key.pem",
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	if _, err := LoadFrom(env(map[string]string{"LOG_LEVEL": "verbose"})); err == nil {
		t.Fatal("expected an error for an invalid LOG_LEVEL")
	}
}

func TestLoadTLSNeedsBothFiles(t *testing.T) {
	for _, m := range []map[string]string{{"TLS_CERT_FILE": "c.pem"}, {"TLS_KEY_FILE": "k.pem"}} {
		if _, err := LoadFrom(env(m)); err == nil {
			t.Errorf("%v: expected an error", m)
		}
	}
}

func TestRequireDatabase(t *testing.T) {
	if err := (Config{}).RequireDatabase(); !errors.Is(err, ErrMissingDatabaseURL) {
		t.Errorf("got %v, want ErrMissingDatabaseURL", err)
	}
	if err := (Config{DatabaseURL: "postgres://x"}).RequireDatabase(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
