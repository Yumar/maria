# S-002: v1 database schema

- **Milestone**: Week 1 — Core
- **FR/NFR**: FR-1, FR-2, FR-13 (groundwork), NFR-1 (groundwork)
- **Status**: todo

## Story
As the developer of PM Agent, I want the full v1 schema (users, settings, goals, tasks, check-ins, reviews, agent actions) created upfront, so that Weeks 2–4 build features without competing migrations and every record is scoped to a user from day one.

## Acceptance criteria
- [ ] Every domain table has `id uuid`, `user_id uuid NOT NULL REFERENCES users`, `created_at` and `updated_at timestamptz`; `updated_at` is maintained by a trigger.
- [ ] Enum-like columns are `text` with `CHECK` constraints that match the spec's values.
- [ ] A task cannot be linked to another user's goal (enforced by a composite foreign key).
- [ ] Migrations apply cleanly up, down and up again.
- [ ] Integration tests prove the constraints above.

## Tasks
| Task | Title | Complexity | Depends on |
|---|---|---|---|
| T-003 | v1 domain schema migration | senior | T-002 |

## Notes
- `text` + `CHECK` constraints are used instead of Postgres `ENUM` types because they are easier to evolve with goose and map cleanly to Go strings in sqlc.
- The `users` table is the tenant root, so it has no `user_id`. It is created in T-002.
- Deviations from and additions to the spec's domain model, all intentional:
  - `goals.cadence_per_week`: the spec's "cadence" is modelled as a target number of tasks done per week (used by FR-4).
  - `tasks.position`: supports reordering (FR-2).
  - `tasks.approval`: models "agent-created tasks start as proposals".
  - `check_ins.goal_id`: links stall check-ins to a goal.
  - `agent_actions.before`, `entity_type`, `entity_id`, `actor_name`: needed to undo changes (FR-13) and to record the MCP client name.
- No data is ever hard-deleted (Agent rule 4), so there are no `ON DELETE CASCADE` rules on domain tables.
- QA: covered by the T-003 integration tests. There is no separate QA pass for this story.
