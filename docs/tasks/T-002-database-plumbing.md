# T-002: Database plumbing: pgx pool, goose, sqlc, testcontainers helper

- **Story**: S-001
- **Issue**: [#6](https://github.com/Yumar/maria/issues/6)
- **FR/NFR**: NFR-3 (groundwork)
- **Complexity**: senior
- **Depends on**: T-001
- **Establishes pattern**: yes (migrations, sqlc queries, transactions, integration tests)

## Goal
Connect to Postgres with pgx, run embedded goose migrations, generate type-safe queries with sqlc, provide a transaction helper, and provide a testcontainers helper for integration tests. Prove the pipeline end to end with the `users` table.

## Files to touch
- `backend/go.mod` — modify. Add deps: `github.com/jackc/pgx/v5`, `github.com/pressly/goose/v3`, `github.com/google/uuid`, `github.com/testcontainers/testcontainers-go` + `.../modules/postgres`. Add tools: `github.com/sqlc-dev/sqlc/cmd/sqlc`, `github.com/pressly/goose/v3/cmd/goose`.
- `backend/internal/db/db.go` — create:
  - `Open(ctx, databaseURL string) (*pgxpool.Pool, error)` (pings before returning)
  - `Migrate(ctx, pool *pgxpool.Pool) error` (goose `Provider` over the embedded FS, via `stdlib.OpenDBFromPool`)
  - `WithTx(ctx, pool, fn func(q *sqlc.Queries) error) error` (commit on nil, rollback on error or panic)
- `backend/internal/db/migrations/embed.go` — create: `//go:embed *.sql` exported as `FS`.
- `backend/internal/db/migrations/00001_users.sql` — create (goose **sequential** numbering):
  - function `set_updated_at()` (trigger: `NEW.updated_at = now()`)
  - table `users(id uuid PK DEFAULT gen_random_uuid(), email text NOT NULL UNIQUE, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now())`
  - trigger `users_set_updated_at BEFORE UPDATE`
  - a Down section
- `backend/internal/db/queries/users.sql` — create: `UpsertUserByEmail :one` (`INSERT … ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING *`), `GetUserByID :one`.
- `backend/sqlc.yaml` — create: engine `postgresql`, `sql_package: pgx/v5`, out `internal/db/sqlc`, package `sqlc`, `emit_pointers_for_null_types: true`, `emit_empty_slices: true`. Overrides: `uuid` → `github.com/google/uuid.UUID` (nullable → `*uuid.UUID`), `timestamptz` → `time.Time`, `date` → `time.Time` (nullable → pointers).
- `backend/internal/db/sqlc/*` — generated, committed.
- `backend/internal/testutil/postgres.go` — create (`//go:build integration`):
  - `Pool(t testing.TB) *pgxpool.Pool`: starts **one** `postgres:17-alpine` container per test binary (`sync.Once`), runs `db.Migrate`, and returns a shared pool
  - `CreateUser(t, pool) uuid.UUID`: inserts a user with a random email, so tests isolate by user instead of truncating tables
- `backend/internal/db/db_integration_test.go` — create (`//go:build integration`): migrations go up→down-to-0→up cleanly; `WithTx` commits on success and rolls back on error; `UpsertUserByEmail` is idempotent; the `updated_at` trigger changes the timestamp on update.
- `backend/internal/config/config.go` (+ test) — modify: add `DATABASE_URL` (required) and `MIGRATE_ON_START` (bool, default `true` in development and `false` in production).
- `backend/cmd/server/main.go` — modify: open the pool, run migrations when `MIGRATE_ON_START`, and close the pool on shutdown.
- `Makefile` — modify:
  - `generate` → `go tool sqlc generate`
  - `check-generated` → generate + `git diff --exit-code`
  - `test-integration` → `go test -race -tags=integration ./...`
  - `migrate-up`, `migrate-down`, `migrate-status` and `migrate-new name=…` (`go tool goose -s -dir internal/db/migrations`)
- `.env.example` — modify: add the new vars (`DATABASE_URL=postgres://maria:maria@localhost:5432/maria?sslmode=disable`).
- `AGENTS.md` — modify the Commands section: add the migrate, generate and integration targets.

## Reference (copy this pattern)
- `backend/internal/config/config.go`, `backend/cmd/server/main.go` from T-001 — config and wiring style.

## Acceptance criteria
- [ ] Server starts against `make db-up` Postgres and applies migrations; restarting is a no-op.
- [ ] `make generate` is deterministic; `make check-generated` passes on a clean tree.
- [ ] Integration helper starts one container per package test binary, not one per test.
- [ ] Integration tests above pass; unit tests don't require Docker.
- [ ] `make lint`, `make test`, `make test-integration` pass.

## Verify with
```bash
make generate && make check-generated
make lint
make test
make test-integration
make db-up && make run   # logs show migrations applied; Ctrl+C
```

## Out of scope
- Domain tables (T-003).
- Postgres advisory-lock scheduler (Week 3).
