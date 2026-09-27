// Package db provides connection, migration and transaction helpers for Postgres.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/Yumar/maria/backend/internal/db/migrations"
	"github.com/Yumar/maria/backend/internal/db/sqlc"
)

// Open creates a pgx connection pool for databaseURL and verifies connectivity.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: open pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return pool, nil
}

// Migrate applies all pending embedded goose migrations to pool.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	conn := stdlib.OpenDBFromPool(pool)
	defer conn.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, conn, migrations.FS)
	if err != nil {
		return fmt.Errorf("db: new migration provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("db: migrate up: %w", err)
	}

	return nil
}

// WithTx runs fn inside a transaction over pool, committing on success and
// rolling back if fn returns an error or panics.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(q *sqlc.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := fn(sqlc.New(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("db: commit tx: %w", err)
	}
	committed = true

	return nil
}
