//go:build integration

package store

import (
	"context"
	"os"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testDatabaseURL is the connection string of the container shared by all
// integration tests in this package.
var testDatabaseURL string

func TestMain(m *testing.M) {
	// The Ryuk reaper container fails to start on Docker Desktop for Windows;
	// the container is terminated explicitly below, so Ryuk is not needed.
	if _, set := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !set && runtime.GOOS == "windows" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("site"),
		postgres.WithUsername("site"),
		postgres.WithPassword("site"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		panic(err)
	}
	testDatabaseURL, err = ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}
	code := m.Run()
	_ = testcontainers.TerminateContainer(ctr)
	os.Exit(code)
}

// openTestStore opens a migrated store against the shared container.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	s, err := Open(ctx, testDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestMigrateCreatesTablesAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	// Second run must be a no-op, and the pool must still be usable after it.
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if err := s.Ping(ctx); err != nil {
		t.Fatalf("ping after migrate: %v", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT table_name FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name <> 'goose_db_version'
		 ORDER BY table_name`)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []string{"cv_education", "cv_experience", "cv_language", "cv_skill_group",
		"main_info", "photo", "place", "social_link", "visit"}
	if !slices.Equal(got, want) {
		t.Errorf("tables = %v, want %v", got, want)
	}
}
