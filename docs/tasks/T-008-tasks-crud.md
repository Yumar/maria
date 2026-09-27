# T-008: Tasks CRUD (copy the goals pattern)

- **Story**: S-004
- **Issue**: [#13](https://github.com/Yumar/maria/issues/13)
- **FR/NFR**: FR-2, FR-13
- **Complexity**: implementer
- **Depends on**: T-007
- **Establishes pattern**: no

## Goal
Implement create, list, get and update for tasks by replicating the goals slice file for file. Completion and reordering are T-009.

## API (add to `api/openapi.yaml`)
- `GET /tasks?goal_id=&status=&due_before=` → 200 `TaskList {items: Task[]}`.
  - Only `approval = 'approved'` tasks.
  - Ordered by `position` asc, then `created_at` asc.
  - `due_before` is a date: return tasks with `due_date < due_before`; tasks without a due date are excluded when this filter is set.
- `POST /tasks` body `TaskCreate` → 201 `Task`.
  - Required: `title` (1–200).
  - Optional: `goal_id` (uuid), `notes` (≤ 5000), `priority` (1–3, default 2), `due_date` (date), `estimate_min` (1–10000).
- `GET /tasks/{taskId}` → 200 | 404.
- `PATCH /tasks/{taskId}` body `TaskUpdate` → 200 | 404 | 422.
  - Optional: `title`, `notes`, `priority`, `status` (`todo|doing|done|skipped`).
  - Nullable (`null` clears): `goal_id`, `due_date`, `estimate_min`.
  - An empty body → 422.
- `Task`: `id`, `goal_id|null`, `title`, `notes`, `status`, `priority`, `due_date|null`, `estimate_min|null`, `completed_at|null`, `source` (`user|agent|mcp`), `approval` (`proposed|approved|rejected`), `position`, `created_at`, `updated_at`. `user_id` is not exposed.

## Business rules (service)
- On create:
  - `source` comes from the context actor: `user` → `user`, `mcp_client` → `mcp`, `agent` → `agent`.
  - `approval` = `approved`.
  - `position` = the user's current max `position` + 1 (0 if none).
- If `goal_id` is set (create or PATCH), it must be a goal of the same user. Otherwise return `apperr.Validation` with field `goal_id` "goal not found". Check it in the service with `GetGoal` scoped by `user_id` (the DB composite FK is only a backstop).
- Status → `done` sets `completed_at = now()`. Leaving `done` clears `completed_at`. Staying `done` keeps the original timestamp.
- Audit tools: `create_task`, `update_task`; entity type `task`.

## Files to touch
- `api/openapi.yaml` — modify: schemas and paths above.
- `backend/internal/httpapi/gen/api.gen.go` — regenerate.
- `backend/internal/db/queries/tasks.sql` — create: `CreateTask`, `GetTask`, `GetTaskForUpdate`, `ListTasks` (`sqlc.narg` filters), `UpdateTask`, `NextTaskPosition`. All filter by `user_id`.
- `backend/internal/db/sqlc/*` — regenerate.
- `backend/internal/tasks/task.go` — create (mirror `goals/goal.go`).
- `backend/internal/tasks/task_test.go` — create (mirror `goals/goal_test.go`, names with `FR2`).
- `backend/internal/tasks/service.go` — create (mirror `goals/service.go`).
- `backend/internal/tasks/service_integration_test.go` — create (mirror the goals tests), plus:
  - `TestTasks_FR2_RejectsOtherUsersGoal`
  - `TestTasks_FR2_ListFilters` (goal, status, due_before)
  - `TestTasks_FR2_CompletedAtFollowsStatus`
  - `TestTasks_FR2_NewTasksAppendPosition`
- `backend/internal/httpapi/tasks.go` — create (mirror `httpapi/goals.go`, with a `TaskService` interface).
- `backend/internal/httpapi/tasks_test.go` — create (mirror `httpapi/goals_test.go`).
- `backend/internal/httpapi/server.go` — modify: add the `tasks TaskService` field and constructor param.
- `backend/cmd/server/main.go` — modify: wire `tasks.NewService(pool)`.

## Reference (copy this pattern)
_Tech Lead: replace with the exact pattern file list reported by T-007._
- `api/openapi.yaml` goals section
- `backend/internal/db/queries/goals.sql`
- `backend/internal/patch/patch.go`
- `backend/internal/goals/goal.go`, `goal_test.go`, `service.go`, `service_integration_test.go`
- `backend/internal/httpapi/goals.go`, `goals_test.go`

## Acceptance criteria
- [ ] Endpoints and business rules above work; responses validate against the spec.
- [ ] Every query filters by `user_id`; another user's task → 404.
- [ ] Each write produces one audit row in the same transaction.
- [ ] Tests listed above pass, following the goals test style.
- [ ] `make check-generated`, `make lint`, `make test`, `make test-integration` pass.

## Verify with
```bash
make generate && make check-generated
make lint
make test
make test-integration
```

## Out of scope
- `complete` and `reorder` endpoints (T-009).
- Proposal approve/reject (Week 3), Today view (FR-3).
- Any change to goals files or shared packages (`patch`, `audit`, `apperr`, `auth`). If they don't fit, stop and report.
