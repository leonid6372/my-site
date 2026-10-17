// Package store owns the PostgreSQL connection pool, the embedded schema
// migrations and all SQL queries.
package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

const (
	maxConns       = 5
	connectTimeout = 30 * time.Second
	maxBackoff     = 2 * time.Second
)

// Store wraps the connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// Open connects to PostgreSQL, retrying with exponential backoff for up to
// 30 s so the API survives starting before the database is ready.
func Open(ctx context.Context, url string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = maxConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	backoff := 250 * time.Millisecond
	for {
		err = pool.Ping(ctx)
		if err == nil {
			return &Store{pool: pool}, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("connect to database: %w", err)
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

// Close closes the pool.
func (s *Store) Close() { s.pool.Close() }

// Ping checks that the database is reachable.
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// Migrate applies all pending embedded migrations. It is a no-op when the
// schema is up to date.
func (s *Store) Migrate(ctx context.Context) error {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	db := stdlib.OpenDBFromPool(s.pool)
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys)
	if err != nil {
		return errors.Join(fmt.Errorf("init migrations: %w", err), db.Close())
	}
	_, err = provider.Up(ctx)
	if err != nil {
		err = fmt.Errorf("migrate: %w", err)
	}
	return errors.Join(err, db.Close())
}
