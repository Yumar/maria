# T-007: Goals CRUD vertical slice (reference pattern)

- **Story**: S-003
- **Issue**: [#12](https://github.com/Yumar/maria/issues/12)
- **FR/NFR**: FR-1, FR-13
- **Complexity**: senior
- **Depends on**: T-005, T-006
- **Establishes pattern**: **yes — the reference for every later CRUD resource and for MCP tools**

## Goal
Implement create, list, get and update (edit + status changes) for goals end to end: OpenAPI → sqlc queries → domain service with audit → strict handlers → tests. Optimize for clarity and ease of copying: implementers will replicate this file-for-file for tasks, check-ins, reviews and settings.

## API (add to `api/openapi.yaml`)
- `GET /goals?status=` → 200 `GoalList {items: Goal[]}`, ordered by `created_at` desc. `status` is an optional enum.
- `POST /goals` body `GoalCreate` → 201 `Goal`.
  - Required: `title` (1–200 chars), `why` (1–2000).
  - Optional: `target_date` (date), `success_metric` (≤ 500), `cadence_per_week` (1–50), `stall_threshold_days` (1–90).
- `GET /goals/{goalId}` → 200 `Goal` | 404.
- `PATCH /goals/{goalId}` body `GoalUpdate` → 200 `Goal` | 404 | 422.
  - All fields optional: `title`, `why`, `success_metric`, `status`.
  - Nullable (`null` clears the value): `target_date`, `cadence_per_week`, `stall_threshold_days`.
  - An empty body (no fields) → 422.
- `Goal`: `id`, `title`, `why`, `target_date|null`, `success_metric`, `cadence_per_week|null`, `stall_threshold_days|null`, `status` (`active|paused|done|dropped`), `created_at`, `updated_at`. `user_id` is **not** exposed.
- There is **no DELETE** (Agent rule 4). "Pause / complete / drop / reactivate" are all `PATCH {status}`, and any transition between the four statuses is allowed so every change stays reversible.

## Files to touch
- `api/openapi.yaml` — modify: the schemas and paths above, reusing the shared `components.responses`.
- `backend/internal/httpapi/gen/api.gen.go` — regenerate.
- `backend/internal/db/queries/goals.sql` — create: `CreateGoal :one`, `GetGoal :one` (`id` + `user_id`), `GetGoalForUpdate :one` (`… FOR UPDATE`), `ListGoals :many` (`user_id`, optional status via `sqlc.narg`), `UpdateGoal :one` (sets all mutable columns; `WHERE id AND user_id`; `RETURNING *`). **Every query filters by `user_id`.**
- `backend/internal/db/sqlc/*` — regenerate.
- `backend/internal/patch/patch.go` — create: a tiny generic type for PATCH semantics, e.g. `Field[T]{Set bool; Null bool; Value T}` with helpers. It is reused by all later PATCH endpoints and MCP `update_*` tools.
- `backend/internal/patch/patch_test.go` — create.
- `backend/internal/goals/goal.go` — create: domain `Goal` struct, `Status` constants, `CreateInput`, `UpdateInput` (using `patch.Field`), `ListFilter`, and pure `validateCreate`/`validateUpdate` that return `apperr.Validation` with field errors.
- `backend/internal/goals/goal_test.go` — create: table-driven validation tests, named with the FR ID (e.g. `TestValidateCreate_FR1`).
- `backend/internal/goals/service.go` — create: `type Service struct{ pool *pgxpool.Pool }`, `NewService`, and the four methods below. The mapping from sqlc rows to domain structs lives here.
  - `List(ctx, userID, ListFilter) ([]Goal, error)`
  - `Create(ctx, userID, CreateInput) (Goal, error)`: validate → `db.WithTx` { insert; `audit.Record` tool `create_goal`, `Before` nil, `After` goal }
  - `Get(ctx, userID, id) (Goal, error)`: `pgx.ErrNoRows` → `apperr.NotFound`
  - `Update(ctx, userID, id, UpdateInput) (Goal, error)`: validate → `WithTx` { `GetGoalForUpdate` (not found → 404) → apply patch in Go → `UpdateGoal` → `audit.Record` tool `update_goal` with `Before`/`After` }. This read-modify-write is the pattern for PATCH, because sqlc `COALESCE` can't clear a value to NULL.
- `backend/internal/goals/service_integration_test.go` — create (`//go:build integration`), using `testutil.Pool` and `testutil.CreateUser`:
  - `TestGoals_FR1_CreateAndGet`
  - `TestGoals_FR1_ListFiltersByStatus`
  - `TestGoals_FR1_UpdateFieldsAndClearNullable`
  - `TestGoals_FR1_StatusTransitions` (pause, complete, drop, reactivate)
  - `TestGoals_UserIsolation`: user B `Get`/`Update` on user A's goal → NotFound
  - `TestGoals_FR13_WritesAreAudited`: rows with the correct tool, actor `user`, before/result
- `backend/internal/httpapi/goals.go` — create: the strict handlers `ListGoals`, `CreateGoal`, `GetGoal`, `UpdateGoal`. They:
  - read the user ID from `auth.UserIDFrom`
  - map between `gen` types and `goals` domain types
  - call a `GoalService` interface declared in this file (the four methods above)
  - return errors as-is (the T-005 error handler maps them)
- `backend/internal/httpapi/goals_test.go` — create: handler unit tests with a fake `GoalService`, covering:
  - 201/200 JSON shape
  - 404 and 422 as problem+json
  - PATCH with `null` vs absent fields reaching the service correctly
  - a schema violation (e.g. `title` too long or bad `status` enum) rejected by the validator with 422
- `backend/internal/httpapi/server.go` — modify: add the `goals GoalService` field and constructor param.
- `backend/cmd/server/main.go` — modify: construct `goals.NewService(pool)` and pass it in.

## Reference (copy this pattern)
- `backend/internal/httpapi/server.go`, `errors.go`, `server_test.go` (T-005): strict handler and error style.
- `backend/internal/audit/audit.go` (T-006), `backend/internal/db/db.go` `WithTx` (T-002).

## Acceptance criteria
- [ ] All endpoints behave as in the API section; responses validate against `api/openapi.yaml`.
- [ ] Every query filters by `user_id`; another user's goal is indistinguishable from a missing one (404).
- [ ] Each create/update writes exactly one `agent_actions` row in the same transaction. A failed write leaves no audit row.
- [ ] Handler unit tests, validation unit tests and service integration tests listed above pass.
- [ ] `make check-generated`, `make lint`, `make test`, `make test-integration` pass.
- [ ] Report lists the **pattern files** so the Tech Lead can record them in S-003 and reference them from T-008.

## Verify with
```bash
make generate && make check-generated
make lint
make test
make test-integration
make db-up && make run &   # then:
curl -s -XPOST localhost:8080/api/v1/goals -H 'content-type: application/json' \
  -d '{"title":"Land a new job","why":"Grow as an engineer","cadence_per_week":5}'
curl -s localhost:8080/api/v1/goals
```

## Out of scope
- Progress summaries on goals (FR-4, Week 3/4) and `list_goals` progress fields.
- Pagination, optimistic concurrency, undo.
- MCP tools (Week 2), which will call `goals.Service` directly.
