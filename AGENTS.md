# AGENTS.md

PM Agent: a personal goal and task tracker with an MCP server and a scheduled accountability agent.
Full spec: `PM Agent — Requirements & Specification.md`. Reference it by section and FR/NFR ID; avoid reading it in full unless planning.

## AI team and workflow

The main session acts as **Tech Lead** (smart model: `opus` or Fusion with Opus lead). Other roles are subagent profiles in `.devin/agents/`:

| Profile | Model | Use for |
|---|---|---|
| `senior-dev` | sonnet | Pattern-setting first slices, complex/cross-cutting work |
| `implementer` | swe | Well-specified tasks that copy an existing reference pattern |
| `reviewer` | gpt | Read-only review of every task's diff before QA/commit |
| `qa` | sonnet | Integration, MCP contract, Playwright, agent evals, FR verification |
| `subagent_explore` (built-in) | default | Codebase/doc research |

Highest-risk work (OAuth 2.1, Postgres advisory-lock scheduler, LLM prompt builder, PWA offline sync) is done by the Tech Lead directly or via `subagent_general` (inherits the main model).

### Flow per milestone
1. **Plan** — Tech Lead writes stories in `docs/stories/` (from `docs/stories/TEMPLATE.md`) and tasks in `docs/tasks/` (from `docs/tasks/TEMPLATE.md`), each tagged with a complexity: `lead`, `senior`, or `implementer`.
2. **Build the pattern first** — the first slice of any new pattern goes to `senior-dev`; later tasks reference its files so `implementer` can copy them.
3. **Dispatch** — run implementers in parallel only when their task file lists do not overlap (watch shared files: migrations, sqlc output, OpenAPI spec, routers).
4. **Review** — `reviewer` on every task. CHANGES REQUESTED goes back to the same agent (resume it) with the findings.
5. **QA** — `qa` verifies the story's FR IDs and adds higher-level tests.
6. **Commit** — Tech Lead commits per task or story, referencing task/FR IDs.

### Task sizing
One task = one sitting: roughly 1–5 files, explicit file list, a reference example when one exists, testable acceptance criteria, and the exact test command.

## Tech stack (from spec)
- **Backend**: Go, chi router, sqlc, goose migrations, PostgreSQL. One binary serves REST API, MCP server and scheduler.
- **MCP**: official MCP Go SDK, Streamable HTTP, OAuth 2.1.
- **Scheduler**: robfig/cron in-process, Postgres advisory lock for single execution.
- **Web**: Angular (standalone components, signals), Angular service worker (PWA + push).
- **API contract**: OpenAPI 3.1, generated TS client.
- **CI**: GitHub Actions (lint, tests, evals, deploy).

## Conventions
- Every record has `id`, `user_id`, `created_at`, `updated_at`; every query filters by `user_id`.
- Agent and MCP writes are logged to `AgentAction` and must be reversible; no destructive MCP tools.
- LLM access only through the Go `LLMProvider` interface.
- Secrets encrypted at rest and never logged.
- Tests reference FR IDs in their names where applicable.
- Developers write unit tests with their code; `qa` owns integration/contract/e2e/evals.

## Repository layout
- `api/` — OpenAPI 3.1 contract (`openapi.yaml`), source of truth for server and client codegen.
- `backend/` — Go module `github.com/yumarxpolanco/maria/backend` (`cmd/server`, `internal/...`).
- `web/` — Angular PWA (planned, Week 4).
- `docs/stories/`, `docs/tasks/` — planning.

## Commands
Run from the repo root (the Makefile delegates to `backend/`). Requires Go (≥ 1.26), golangci-lint v2, and Docker (Colima works).
- `make build` — build `backend/bin/server`
- `make run` — run the server locally (config via env vars; see `.env.example`)
- `make fmt` / `make lint` — format / lint (golangci-lint)
- `make test` — unit tests, no Docker
- `make test-integration` — integration tests with testcontainers (starts a `postgres:17-alpine` container per package test binary)
- `make generate` / `make check-generated` — sqlc codegen and drift check
- `make db-up` / `make db-down` — local Postgres 17 via Docker Compose
- `make migrate-up` / `make migrate-down` / `make migrate-status` — apply/roll back/inspect goose migrations against `DATABASE_URL`
- `make migrate-new name=add_foo` — create a new sequential goose migration file
- `make help` — list all targets
_E2E and eval commands: TBD (Weeks 3–4)._
