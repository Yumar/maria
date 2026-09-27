# T-006: Audit recorder (AgentAction)

- **Story**: S-003
- **Issue**: [#11](https://github.com/Yumar/maria/issues/11)
- **FR/NFR**: FR-13, FR-14 (groundwork)
- **Complexity**: senior
- **Depends on**: T-003
- **Establishes pattern**: yes (every service write calls `audit.Record` inside its transaction)

## Goal
Provide one small API that services call inside their write transaction to log what changed, who changed it, and the state before and after. REST writes are logged as `actor=user` today; Week 2 MCP and the Week 3 agent only set a different actor in the context.

## Files to touch
- `backend/internal/audit/audit.go` — create:
  ```go
  type ActorKind string // ActorUser = "user", ActorAgent = "agent", ActorMCPClient = "mcp_client"
  type Actor struct { Kind ActorKind; Name string } // Name: MCP client name, "" for user
  func WithActor(ctx context.Context, a Actor) context.Context
  func ActorFrom(ctx context.Context) Actor // default Actor{Kind: ActorUser}
  type Entry struct {
      UserID     uuid.UUID
      Tool       string     // MCP-style tool name, e.g. "create_goal", "update_task"
      EntityType string     // "goal", "task", ...
      EntityID   *uuid.UUID // nil for multi-entity actions (e.g. reorder)
      Input      any        // request input as received (JSON-marshalled)
      Before     any        // nil on create
      After      any        // stored in the `result` column
  }
  func Record(ctx context.Context, q *sqlc.Queries, e Entry) error
  ```
  `Record` marshals `Input`, `Before` and `After` to JSON (`Before`/`After` nil → SQL NULL; `Input` nil → `{}`), takes the actor from the ctx, and inserts. It must be called with the `*sqlc.Queries` bound to the caller's transaction (`db.WithTx`), so the audit row commits or rolls back together with the change.
- `backend/internal/db/queries/agent_actions.sql` — create: `InsertAgentAction :one`, `ListAgentActions :many` (by `user_id`, newest first, `LIMIT` param), `GetAgentAction :one` (by `id` **and** `user_id`).
- `backend/internal/db/sqlc/*` — regenerate.
- `backend/internal/audit/audit_test.go` — create (unit): actor context default and override; JSON marshalling helper (split the marshalling into a pure function so it's unit-testable).
- `backend/internal/audit/audit_integration_test.go` — create (`//go:build integration`):
  - `TestAudit_FR13_RecordInsideTx`: `Record` inside a `WithTx` that commits → row exists with the right actor/tool/before/result
  - rollback → no row
  - `ActorMCPClient` with a name is stored in `actor`/`actor_name`

## Reference (copy this pattern)
- `backend/internal/db/db.go` (`WithTx`), `backend/internal/db/queries/users.sql`, `backend/internal/db/db_integration_test.go` (T-002).

## Acceptance criteria
- [ ] API matches the signatures above (T-007 depends on them).
- [ ] Secrets never end up in audit rows. Add a doc comment on `Entry.Input` stating callers must not pass secrets. Nothing secret exists in Week 1, but settings (API key) will later.
- [ ] Unit and integration tests above pass; `make check-generated`, `make lint`, `make test`, `make test-integration` pass.

## Verify with
```bash
make generate && make check-generated
make lint
make test
make test-integration
```

## Out of scope
- Undo/revert logic and the history endpoint (Week 2 / Week 4, FR-13 UI).
- Autonomy checks turning writes into proposals (Week 3, FR-14).
- Do not touch `internal/httpapi` or `api/openapi.yaml`; T-005 may be running in parallel.
