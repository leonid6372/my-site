//go:build integration

package main

import (
	"context"
	"errors"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/leonid6372/my-site/backend/internal/config"
	"github.com/leonid6372/my-site/backend/internal/store"
)

// emptyDB starts a fresh PostgreSQL without any schema and returns its URL.
func emptyDB(t *testing.T) string {
	t.Helper()
	// Ryuk does not start on Docker Desktop for Windows; the container is
	// terminated in t.Cleanup instead.
	if _, set := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !set && runtime.GOOS == "windows" {
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("site"), postgres.WithUsername("site"), postgres.WithPassword("site"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(ctr) })
	url, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return url
}

// FR-B10: serve's startup path creates the schema on an empty database,
// adds no content, and migrating again is a no-op.
func TestMigrateEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{DatabaseURL: emptyDB(t)}

	s, err := openMigrated(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.GetMain(ctx); !errors.Is(err, store.ErrNoContent) {
		t.Errorf("GetMain err = %v, want ErrNoContent (startup must not add data)", err)
	}
	if places, err := s.ListPlaces(ctx); err != nil || len(places) != 0 {
		t.Errorf("ListPlaces = %d places, %v; want none", len(places), err)
	}

	s2, err := openMigrated(ctx, cfg)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	s2.Close()
}
