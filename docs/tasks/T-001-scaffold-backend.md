# T-001: Scaffold Go backend and dev tooling

- **Story**: S-001
- **FR/NFR**: NFR-6
- **Complexity**: senior
- **Depends on**: —
- **Establishes pattern**: yes (project layout, config, logging, Makefile conventions)

## Goal
Create the Go module, an HTTP server with a health check, structured logging, graceful shutdown, local Postgres via Docker Compose, lint config and a Makefile. This is the skeleton every later task plugs into.

## Files to touch
- `backend/go.mod` — create: module `github.com/Yumar/maria/backend`, latest stable Go (≥ 1.26). Dependency: `github.com/go-chi/chi/v5`.
- `backend/cmd/server/main.go` — create: load config → build `slog` logger (JSON; text when `APP_ENV=development` and `LOG_FORMAT=text`) → build router → `http.Server` with read/write/idle timeouts → graceful shutdown on SIGINT/SIGTERM (10 s).
- `backend/internal/config/config.go` — create: `Config` struct and `Load() (Config, error)` using only stdlib `os.Getenv` with defaults. Vars: `APP_ENV` (`development`|`production`, default `development`), `HTTP_ADDR` (default `:8080`), `LOG_LEVEL` (`debug|info|warn|error`, default `info`), `LOG_FORMAT` (`json|text`, default `json`). Return an error for invalid values.
- `backend/internal/config/config_test.go` — create: table-driven tests for defaults, overrides and invalid values (use `t.Setenv`).
- `backend/internal/httpapi/router.go` — create: `NewRouter(logger *slog.Logger) http.Handler` with chi middleware `RequestID`, `Recoverer` (`RealIP` was dropped: deprecated in chi v5.3 because it allows IP spoofing), a request-logging middleware (method, path, status, duration_ms, request_id; **never** bodies or headers), and `GET /healthz` → `200 {"status":"ok"}`.
- `backend/internal/httpapi/router_test.go` — create: `httptest` test for `/healthz`, and for 404 on unknown routes.
- `backend/.golangci.yml` — create: golangci-lint v2 config. Linters: defaults + `gosec`, `bodyclose`, `errorlint`, `gocritic`, `misspell`, `revive`. Formatters: `gofmt`, `goimports` (local prefix `github.com/Yumar/maria`). Exclude generated files.
- `Makefile` — create (repo root; each target `cd backend`): `help`, `build` (→ `backend/bin/server`), `run`, `fmt`, `lint`, `test` (`go test -race ./...`), `db-up`, `db-down` (docker compose). Placeholder targets `generate`, `check-generated` and `test-integration` that exit 0 with "not configured yet".
- `docker-compose.yml` — create: `postgres:17-alpine`, db/user/password `maria`, port `5432`, named volume, `pg_isready` healthcheck.
- `.env.example` — create: every config var with its default, with comments.
- `.gitignore` — create: `backend/bin/`, `.env`, OS/IDE files, coverage output.
- `.editorconfig` — create: tabs for Go and Makefile, 2 spaces for YAML/JSON/TS/MD.
- `AGENTS.md` — modify: replace "Commands" TBD with the real `make` targets and add a short "Repository layout" section (`api/`, `backend/`, `web/` planned).

## Reference (copy this pattern)
None — this task establishes the pattern. Keep `main.go` thin: it wires things together, and all logic lives in `internal/`.

## Acceptance criteria
- [ ] `make build` produces `backend/bin/server`; `make run` serves `GET /healthz` → 200 `{"status":"ok"}`.
- [ ] SIGINT shuts the server down cleanly (logs "shutdown complete"; exit code 0).
- [ ] Request logs are JSON by default and contain `method`, `path`, `status`, `duration_ms` and `request_id`.
- [ ] Invalid `APP_ENV` or `LOG_LEVEL` makes startup fail with a clear error.
- [ ] `make db-up` starts a healthy Postgres container.
- [ ] Unit tests cover config parsing and the health/404 routes.
- [ ] `make lint` and `make test` pass.

## Verify with
```bash
make lint
make test
make build && (./backend/bin/server & sleep 1; curl -fsS localhost:8080/healthz; kill %1)
make db-up && docker compose ps
```

## Out of scope
- Database connection, migrations and sqlc (T-002).
- OpenAPI, auth and error model (T-004, T-005).
- CI workflow (T-010).
- Dockerfile for deployment (Week 4).
