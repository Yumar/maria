# T-003: v1 domain schema migration

- **Story**: S-002
- **FR/NFR**: FR-1, FR-2, FR-13, NFR-1
- **Complexity**: senior
- **Depends on**: T-002
- **Establishes pattern**: no (follows the T-002 migration style)

## Goal
Create all remaining v1 tables in one migration so later milestones don't compete for migration files. Add integration tests proving the key constraints.

## Files to touch
- `backend/internal/db/migrations/00002_domain_schema.sql` — create (Up and Down).
- `backend/internal/db/schema_integration_test.go` — create (`//go:build integration`).

## Schema
All tables share these columns and rules:
- Common columns: `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`, `user_id uuid NOT NULL REFERENCES users(id)`, `created_at timestamptz NOT NULL DEFAULT now()`, `updated_at timestamptz NOT NULL DEFAULT now()`.
- A `BEFORE UPDATE` trigger using `set_updated_at()` on every table.
- No `ON DELETE CASCADE`.

**user_settings** (one row per user; `UNIQUE (user_id)`)
- `tone text NOT NULL DEFAULT 'direct'` CHECK in (`coach`, `direct`, `drill_sergeant`)
- `autonomy text NOT NULL DEFAULT 'suggest_only'` CHECK in (`suggest_only`, `act_and_log`)
- `checkin_time time NOT NULL DEFAULT '08:30'`
- `review_weekday smallint NOT NULL DEFAULT 0` CHECK 0–6 (0 = Sunday); `review_time time NOT NULL DEFAULT '18:00'`
- `quiet_start time NOT NULL DEFAULT '21:00'`, `quiet_end time NOT NULL DEFAULT '07:30'`
- `daily_nudge_cap int NOT NULL DEFAULT 3` CHECK ≥ 0
- `default_stall_threshold_days int NOT NULL DEFAULT 3` CHECK 1–90
- `channels text[] NOT NULL DEFAULT '{push}'` CHECK `channels <@ ARRAY['push','email']`
- `email_fallback_delay_min int NOT NULL DEFAULT 120` CHECK ≥ 0
- `llm_provider text NOT NULL DEFAULT 'anthropic'`; `llm_model text NOT NULL DEFAULT ''` (empty means the app default)
- `llm_api_key_encrypted bytea NULL` (encrypted by the app, never plaintext)
- `language text NOT NULL DEFAULT 'en'` CHECK in (`en`, `es`)
- `timezone text NOT NULL DEFAULT 'America/Santo_Domingo'`

**goals**
- `title text NOT NULL` CHECK `length(btrim(title)) > 0`; `why text NOT NULL` CHECK `length(btrim(why)) > 0`
- `target_date date NULL`, `success_metric text NOT NULL DEFAULT ''`
- `cadence_per_week smallint NULL` CHECK 1–50 (target tasks done per week)
- `stall_threshold_days smallint NULL` CHECK 1–90 (NULL means use the user default)
- `status text NOT NULL DEFAULT 'active'` CHECK in (`active`, `paused`, `done`, `dropped`)
- `UNIQUE (id, user_id)` (target of the composite FK); index `(user_id, status)`

**tasks**
- `goal_id uuid NULL`, with `FOREIGN KEY (goal_id, user_id) REFERENCES goals(id, user_id)` so a task can't link to another user's goal
- `title text NOT NULL` CHECK non-blank; `notes text NOT NULL DEFAULT ''`
- `status text NOT NULL DEFAULT 'todo'` CHECK in (`todo`, `doing`, `done`, `skipped`)
- `priority smallint NOT NULL DEFAULT 2` CHECK 1–3
- `due_date date NULL`, `estimate_min int NULL` CHECK > 0, `completed_at timestamptz NULL`
- `source text NOT NULL DEFAULT 'user'` CHECK in (`user`, `agent`, `mcp`)
- `approval text NOT NULL DEFAULT 'approved'` CHECK in (`proposed`, `approved`, `rejected`)
- `position int NOT NULL DEFAULT 0`
- Indexes: `(user_id, status)`, `(user_id, goal_id)`, `(user_id, due_date)`, `(user_id, position)`

**check_ins**
- `date date NOT NULL`, `kind text NOT NULL` CHECK in (`daily`, `stall`, `ad_hoc`)
- `goal_id uuid NULL` with the same composite FK as tasks
- `done_summary text NOT NULL DEFAULT ''`, `blockers text NOT NULL DEFAULT ''`
- `energy smallint NULL` CHECK 1–5; `agent_reply text NOT NULL DEFAULT ''`
- Index `(user_id, date DESC)`

**reviews**
- `week_start date NOT NULL`, `goal_progress jsonb NOT NULL DEFAULT '[]'`
- `wins text NOT NULL DEFAULT ''`, `slips text NOT NULL DEFAULT ''`, `plan_next_week text NOT NULL DEFAULT ''`
- `UNIQUE (user_id, week_start)`

**agent_actions**
- `actor text NOT NULL` CHECK in (`agent`, `mcp_client`, `user`); `actor_name text NOT NULL DEFAULT ''`
- `tool text NOT NULL`, `entity_type text NOT NULL`, `entity_id uuid NULL`
- `input jsonb NOT NULL DEFAULT '{}'`, `before jsonb NULL`, `result jsonb NULL`
- `reverted_at timestamptz NULL`
- Index `(user_id, created_at DESC)`

## Reference (copy this pattern)
- `backend/internal/db/migrations/00001_users.sql` — goose annotations, trigger style, Down section.
- `backend/internal/db/db_integration_test.go`, `backend/internal/testutil/postgres.go` — integration test style.

## Acceptance criteria
- [ ] Migration applies up → down → up cleanly.
- [ ] Integration tests (named with FR IDs where applicable) prove:
  - an invalid enum value is rejected (e.g. goal `status='x'`)
  - `TestSchema_FR2_TaskCannotLinkOtherUsersGoal`: inserting a task with user B and user A's `goal_id` fails
  - `priority` 0 and 4 are rejected
  - `user_settings` is unique per user
  - `updated_at` changes on update for `goals`
- [ ] `make check-generated` still passes (no queries added, so sqlc output is unchanged).
- [ ] `make lint`, `make test`, `make test-integration` pass.

## Verify with
```bash
make check-generated
make lint
make test
make test-integration
```

## Out of scope
- sqlc queries for these tables (added by each feature task).
- Seeding settings rows (Week 4 settings feature).
