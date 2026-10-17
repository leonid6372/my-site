//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/leonid6372/my-site/backend/internal/config"
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

// brokenFixture copies the valid fixture with one empty ru text.
func brokenFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join("..", "..", "internal", "content", "testdata", "valid")
	for _, name := range []string{"main.yaml", "cv.yaml", "travel.yaml"} {
		data, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "travel.yaml" {
			data = bytes.Replace(data, []byte(`ru: "Кито"`), []byte(`ru: ""`), 1)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// FR-B07, FR-B08, FR-B10: the seed path migrates an empty database, imports
// the content, and an invalid content directory changes nothing.
func TestSeedCommandPath(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{DatabaseURL: emptyDB(t)}

	s, err := openMigrated(ctx, cfg) // FR-B10: schema created on an empty DB
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	repoContent := filepath.Join("..", "..", "..", "content")
	if err := seed(ctx, s, repoContent); err != nil {
		t.Fatalf("seed repo content: %v", err)
	}
	before, err := s.ListPlaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	mainBefore, err := s.GetMain(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) == 0 {
		t.Fatal("no places after seeding the repo content")
	}

	// FR-B08: invalid content → error, database untouched.
	if err := seed(ctx, s, brokenFixture(t)); err == nil {
		t.Fatal("seed of invalid content succeeded")
	}
	after, err := s.ListPlaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	mainAfter, err := s.GetMain(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) || mainAfter.FullName != mainBefore.FullName {
		t.Errorf("invalid seed changed the database: %d → %d places, %q → %q",
			len(before), len(after), mainBefore.FullName.EN, mainAfter.FullName.EN)
	}

	// Migrating again on the seeded database is a no-op and keeps the data.
	s2, err := openMigrated(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	s2.Close()
	if again, err := s.ListPlaces(ctx); err != nil || len(again) != len(before) {
		t.Errorf("re-migrate: %d places, %v", len(again), err)
	}
}
