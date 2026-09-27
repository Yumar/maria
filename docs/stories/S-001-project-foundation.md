# S-001: Project foundation

- **Milestone**: Week 1 — Core
- **FR/NFR**: NFR-3 (groundwork), NFR-6
- **Status**: todo

## Story
As the developer of PM Agent, I want a runnable Go backend with Postgres, migrations, code generation, lint, tests and CI, so that every later feature ships on a consistent, verified base.

## Acceptance criteria
- [ ] `make db-up && make run` starts the server on `:8080`; `GET /healthz` returns `200 {"status":"ok"}`.
- [ ] Migrations run automatically on startup in development and via `make migrate-up`.
- [ ] `make generate` regenerates sqlc code (and OpenAPI code once S-003 lands); `make check-generated` fails when generated code is stale.
- [ ] `make lint`, `make test` (no Docker) and `make test-integration` (testcontainers) all pass.
- [ ] Logs are structured JSON (`log/slog`); request logs include method, path, status, duration and request ID, and never bodies or auth headers.
- [ ] GitHub Actions runs lint, unit tests, integration tests and the generated-code drift check on every push and PR.
- [ ] `AGENTS.md` "Commands" section lists the real commands.

## Tasks
| Task | Title | Complexity | Depends on |
|---|---|---|---|
| T-001 | Scaffold Go backend and dev tooling | senior | — |
| T-002 | Database plumbing: pgx pool, goose, sqlc, testcontainers helper | senior | T-001 |
| T-010 | CI workflow (GitHub Actions) | implementer | T-005 |

## Notes
- **Repo layout** (monorepo): `api/openapi.yaml` (contract shared with the Angular client in Week 4), `backend/` (Go module `github.com/Yumar/maria/backend`), `web/` (Week 4). Root `Makefile` delegates to `backend/`.
- **Go version**: latest stable (≥ 1.26). oapi-codegen v2.8+ needs a recent toolchain.
- **Code generators** are pinned as Go tools (`go get -tool`) and run with `go tool sqlc`, `go tool oapi-codegen` and `go tool goose`. golangci-lint is installed as a binary (its recommended install method), not as a Go tool.
- **Test split**: `make test` runs unit tests only. DB-backed tests use the `//go:build integration` tag and run with `make test-integration`.
- Week 1 dispatch order (the `∥` symbol means the tasks can run in parallel):
  `T-001 → T-002 → (T-003 ∥ T-004) → (T-005 ∥ T-006) → T-007 → T-008 → T-009`, with `T-010` any time after T-005.
