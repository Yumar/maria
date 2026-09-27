# T-009: Complete and reorder tasks

- **Story**: S-004
- **FR/NFR**: FR-2, FR-13
- **Complexity**: implementer
- **Depends on**: T-008
- **Establishes pattern**: no

## Goal
Add the two non-CRUD task operations: completing a task (optionally with a note) and reordering a set of tasks. These map directly to the Week 2 MCP tool `complete_task` and to drag-and-drop in the Week 4 UI.

## API (add to `api/openapi.yaml`)
- `POST /tasks/{taskId}/complete`, body `TaskComplete {note?: string ≤ 2000}` (the body is optional) → 200 `Task` | 404.
- `POST /tasks/reorder`, body `TaskReorder {task_ids: uuid[]}` (min 2, max 500, `uniqueItems: true`) → 204 | 422.

## Business rules (service)
**Complete** (`Service.Complete(ctx, userID, id, note string) (Task, error)`), in one transaction:
1. `GetTaskForUpdate` (not found → `apperr.NotFound`).
2. If it's already `done` and `note` is empty, return the task unchanged: no update, no audit row. If it's already `done` and `note` is non-empty, only append the note.
3. Otherwise set `status = done` and `completed_at = now()`.
4. A non-empty note is appended to `notes`: if `notes` is non-empty, add `"\n\n"` + note; otherwise the note alone.
5. `UpdateTask`, then `audit.Record` with tool `complete_task`, input `{note}`, and before/after.

**Reorder** (`Service.Reorder(ctx, userID, ids []uuid.UUID) error`), in one transaction:
1. Load the listed tasks with `SELECT id, position FROM tasks WHERE user_id = $1 AND id = ANY($2) FOR UPDATE`.
2. If the number of rows found ≠ `len(ids)` → `apperr.Validation` field `task_ids` "unknown task id". Duplicates are also a validation error (checked in Go even though the schema enforces `uniqueItems`).
3. Take the **existing** positions of those tasks, sort them ascending, and assign them in the order of `ids` (`ids[i]` gets `sorted[i]`). Tasks not in the list are untouched, so reordering a filtered view (e.g. one goal) never disturbs other tasks.
   - If sorted positions contain duplicates (possible with legacy data), assign `sorted[0] + i` instead.
4. Update each position with `UpdateTaskPosition` (`WHERE id AND user_id`).
5. One `audit.Record`:
   - tool `reorder_tasks`, entity type `task`, `EntityID` nil
   - input `{task_ids}`
   - before/after as `map[taskID]position`

## Files to touch
- `api/openapi.yaml` — modify: the two paths and schemas above.
- `backend/internal/httpapi/gen/api.gen.go` — regenerate.
- `backend/internal/db/queries/tasks.sql` — modify: add `LockTasksForReorder :many` and `UpdateTaskPosition :exec`.
- `backend/internal/db/sqlc/*` — regenerate.
- `backend/internal/tasks/service.go` — modify: add `Complete` and `Reorder`. Extract the position-assignment logic as a pure function (e.g. `assignPositions(ids, current map[uuid.UUID]int) map[uuid.UUID]int`).
- `backend/internal/tasks/task_test.go` — modify: unit tests for `assignPositions`, covering subset reorder, reversed order, and duplicate positions.
- `backend/internal/tasks/service_integration_test.go` — modify:
  - `TestTasks_FR2_Complete`
  - `TestTasks_FR2_CompleteIsIdempotent`
  - `TestTasks_FR2_CompleteAppendsNote`
  - `TestTasks_FR2_ReorderSubsetLeavesOthersUntouched`
  - `TestTasks_FR2_ReorderRejectsOtherUsersTask`
  - `TestTasks_FR13_CompleteAndReorderAudited`
- `backend/internal/httpapi/tasks.go` — modify: the two handlers; extend the `TaskService` interface.
- `backend/internal/httpapi/tasks_test.go` — modify: handler tests for 200/204/404/422.

## Reference (copy this pattern)
- `backend/internal/tasks/service.go` `Update` (T-008) and `backend/internal/goals/service.go` `Update` (T-007): transaction + lock + audit style.
- `backend/internal/httpapi/tasks.go` / `tasks_test.go` (T-008).

## Acceptance criteria
- [ ] Complete and reorder behave exactly per the rules above.
- [ ] All queries filter by `user_id`; a foreign task ID in reorder → 422, not a partial update.
- [ ] Unit and integration tests listed above pass.
- [ ] `make check-generated`, `make lint`, `make test`, `make test-integration` pass.

## Verify with
```bash
make generate && make check-generated
make lint
make test
make test-integration
```

## Out of scope
- Changes to create/update semantics from T-008.
- Undo of reorder/complete (FR-13 UI, later).
