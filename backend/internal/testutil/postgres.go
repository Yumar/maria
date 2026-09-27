//go:build integration

// Package testutil provides shared helpers for integration tests.
package testutil

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Yumar/maria/backend/internal/db"
	"github.com/Yumar/maria/backend/internal/db/sqlc"
)

var (
	once       sync.Once
	sharedPool *pgxpool.Pool
	initErr    error
)

// Pool starts one postgres:17-alpine container per test binary (shared across
// all tests via sync.Once), runs migrations, and returns the shared pool.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()

	once.Do(func() {
		ctx := context.Background()

		container, err := postgres.Run(ctx, "postgres:17-alpine",
			postgres.WithDatabase("maria"),
			postgres.WithUsername("maria"),
			postgres.WithPassword("maria"),
			postgres.BasicWaitStrategies(),
		)
		if err != nil {
			initErr = fmt.Errorf("testutil: start postgres container: %w", err)
			return
		}

		connStr, err := container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			initErr = fmt.Errorf("testutil: connection string: %w", err)
			return
		}

		pool, err := db.Open(ctx, connStr)
		if err != nil {
			initErr = fmt.Errorf("testutil: open pool: %w", err)
			return
		}

		if err := db.Migrate(ctx, pool); err != nil {
			initErr = fmt.Errorf("testutil: migrate: %w", err)
			return
		}

		sharedPool = pool
	})

	if initErr != nil {
		t.Fatalf("testutil: setup postgres: %v", initErr)
	}

	return sharedPool
}

// CreateUser inserts a user with a random email and returns its ID. Tests
// isolate by user rather than truncating shared tables.
func CreateUser(t testing.TB, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	q := sqlc.New(pool)

	email := fmt.Sprintf("%s@example.test", uuid.NewString())
	user, err := q.UpsertUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("testutil: create user: %v", err)
	}

	return user.ID
}
