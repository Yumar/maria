# S-003: API foundation and Goals CRUD

- **Milestone**: Week 1 — Core
- **FR/NFR**: FR-1, FR-13 (audit recording), NFR-4
- **Status**: todo

## Story
As the PM Agent user, I want to create, view, edit, pause, complete and drop goals through a documented REST API, so that my goals live in one place that the web app and (in Week 2) MCP clients can share.

## Acceptance criteria
- [ ] `api/openapi.yaml` is a valid OpenAPI 3.1 document; the Go server types are generated from it with oapi-codegen (strict server on chi).
- [ ] Every `/api/v1` request is authenticated. In Week 1 a dev stub resolves a single seeded user; the stub refuses to start when `APP_ENV=production`.
- [ ] Errors are returned as RFC 9457 `application/problem+json`: 400 malformed body, 401 unauthenticated, 404 not found (including another user's records), 422 validation errors with per-field details.
- [ ] `GET/POST /api/v1/goals` and `GET/PATCH /api/v1/goals/{goalId}` work as specified in T-007; `status` supports active ↔ paused ↔ done ↔ dropped. There is no DELETE endpoint.
- [ ] Every goal write records an `agent_actions` row in the same transaction, with the actor taken from the request context (`user` for REST), the tool name, the input, and the before/after state.
- [ ] All queries filter by `user_id`.

## Tasks
| Task | Title | Complexity | Depends on |
|---|---|---|---|
| T-004 | Error model and dev auth stub | implementer | T-002 |
| T-005 | OpenAPI foundation, codegen and router wiring | senior | T-004 |
| T-006 | Audit recorder (AgentAction) | senior | T-003 |
| T-007 | Goals CRUD vertical slice (**reference pattern**) | senior | T-005, T-006 |

## Notes
- T-007 is **the** reference pattern for all later CRUD work (tasks, check-ins, reviews, settings) and for MCP tools, which will call the same services. When it's done, record its file list here:
  - _pattern files: (fill in after T-007 review)_
- Layering: `handler (httpapi) → service (internal/<domain>) → sqlc queries`. Handlers depend on a small interface defined in `httpapi` so they can be unit-tested with fakes. Services take `userID` explicitly.
- Base path `/api/v1`; `/healthz` stays outside the versioned API.
- Lists return `{ "items": [...] }` without pagination (one user, small data). The wrapper leaves room to add pagination later without breaking clients.
- **QA** (after T-007): HTTP-level integration tests against the real router and Postgres covering FR-1 end to end, user isolation (user B gets 404 for user A's goal), audit rows for each write, and response bodies validated against `api/openapi.yaml`.
