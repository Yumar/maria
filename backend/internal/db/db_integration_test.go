//go:build integration

package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/yumarxpolanco/maria/backend/internal/db"
	"github.com/yumarxpolanco/maria/backend/internal/db/migrations"
	"github.com/yumarxpolanco/maria/backend/internal/db/sqlc"
	"github.com/yumarxpolanco/maria/backend/internal/testutil"
)

func TestMigrate_UpDownUp(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()

	conn := stdlib.OpenDBFromPool(pool)
	defer conn.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, conn, migrations.FS)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("Up() error = %v", err)
	}

	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("DownTo(0) error = %v", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("second Up() error = %v", err)
	}
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()

	email := "with-tx-commit@example.test"
	var gotID any

	err := db.WithTx(ctx, pool, func(q *sqlc.Queries) error {
		user, err := q.UpsertUserByEmail(ctx, email)
		if err != nil {
			return err
		}
		gotID = user.ID
		return nil
	})
	if err != nil {
		t.Fatalf("WithTx() error = %v", err)
	}

	count := countByEmail(t, pool, email)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if gotID == nil {
		t.Errorf("expected a user ID to be captured")
	}
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()

	email := "with-tx-rollback@example.test"
	wantErr := errors.New("boom")

	err := db.WithTx(ctx, pool, func(q *sqlc.Queries) error {
		if _, err := q.UpsertUserByEmail(ctx, email); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx() error = %v, want %v", err, wantErr)
	}

	if count := countByEmail(t, pool, email); count != 0 {
		t.Errorf("count = %d, want 0 (rolled back)", count)
	}
}

func TestUpsertUserByEmail_Idempotent(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	q := sqlc.New(pool)

	email := "idempotent@example.test"

	first, err := q.UpsertUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("UpsertUserByEmail() error = %v", err)
	}

	second, err := q.UpsertUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("UpsertUserByEmail() second call error = %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("ID changed across upserts: %v != %v", first.ID, second.ID)
	}
	if count := countByEmail(t, pool, email); count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestUpdatedAtTrigger_ChangesOnUpdate(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()
	q := sqlc.New(pool)

	email := "trigger@example.test"

	created, err := q.UpsertUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("UpsertUserByEmail() error = %v", err)
	}

	updated, err := q.UpsertUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("UpsertUserByEmail() second call error = %v", err)
	}

	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want after %v", updated.UpdatedAt, created.UpdatedAt)
	}
}

func countByEmail(t *testing.T, pool *pgxpool.Pool, email string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE email = $1", email).Scan(&count); err != nil {
		t.Fatalf("countByEmail() error = %v", err)
	}
	return count
}
