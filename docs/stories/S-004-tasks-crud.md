# S-004: Tasks CRUD, completion and reordering

- **Milestone**: Week 1 — Core
- **FR/NFR**: FR-2, FR-13 (audit recording)
- **Status**: todo

## Story
As the PM Agent user, I want to create, edit, reorder and complete tasks, optionally linked to a goal, so that each goal turns into concrete next actions I can track.

## Acceptance criteria
- [ ] `GET/POST /api/v1/tasks` and `GET/PATCH /api/v1/tasks/{taskId}` work as specified in T-008. The list filters by `goal_id`, `status` and `due_before`.
- [ ] Linking a task to a goal that doesn't exist or belongs to another user returns 422 on the `goal_id` field.
- [ ] `POST /api/v1/tasks/{taskId}/complete` marks the task done, sets `completed_at` and optionally appends a note. Completing an already-done task is a no-op.
- [ ] Moving a task to `done` by any route sets `completed_at`; moving it out of `done` clears it.
- [ ] `POST /api/v1/tasks/reorder` reorders the given tasks among their own current positions without moving any other task.
- [ ] Every write records an `agent_actions` row (tools `create_task`, `update_task`, `complete_task`, `reorder_tasks`).
- [ ] REST-created tasks have `source=user` and `approval=approved`. Lists exclude proposals (`approval <> 'approved'`) by default.

## Tasks
| Task | Title | Complexity | Depends on |
|---|---|---|---|
| T-008 | Tasks CRUD (copy the goals pattern) | implementer | T-007 |
| T-009 | Complete and reorder tasks | implementer | T-008 |

## Notes
- T-008 and T-009 touch the same files, so they run one after the other. Resume the same implementer for T-009 so it keeps its context.
- Proposal approve/reject endpoints (FR-5) are Week 3 scope. The `approval` column already exists.
- **QA** (after T-009): HTTP-level integration tests for FR-2 (CRUD, filters, cross-user `goal_id` rejection, completion idempotency, reorder in a filtered subset leaves other tasks untouched, audit rows), with responses validated against `api/openapi.yaml`. Then extend the S-003 suite into one "create goal → create tasks → complete task" API flow.
