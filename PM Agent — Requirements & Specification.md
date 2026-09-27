# PM Agent — Requirements & Specification

Sep 26, 2026 · @Yumar

## Overview

PM Agent is a personal goal and task tracker whose agent notices drift and pushes you to take the next concrete step. Its data is exposed through an MCP server, so any MCP-capable assistant can read and update it, while its own scheduled agent handles proactive check-ins.

**Problem.** Todo apps store tasks but never follow up. Goals stall silently, and nobody asks why.

**Goals**

- Turn each goal into concrete next actions and keep them moving.
- Reach out proactively: daily check-ins, stall alerts, weekly reviews.
- Work from any assistant via MCP, not just a proprietary chat.
- Serve as a portfolio project showing Go, Angular, MCP and agent design.

**Non-goals (v1)**

- Team or shared workspaces.
- Native mobile apps (PWA covers mobile).
- WhatsApp or Telegram messaging.
- Calendar sync and third-party task imports.

## Scope

v1 is a single-user product: Go backend, MCP server, scheduled agent and an Angular PWA. Flutter and richer channels come later.

| Phase | Includes |
| --- | --- |
| v1 (MVP) | Goals and tasks CRUD, MCP server, daily check-in, stall detection, weekly review, PWA push, settings |
| v2 | Flutter mobile app, OpenAPI-generated Dart and TS clients, calendar sync, agent memory summaries |
| v3 | Multi-user and shared goals, WhatsApp channel, integrations (Todoist, GitHub, Notion) |

The data model carries a `user_id` on every record from day one, so v3 needs no migration of existing data.

## Domain model

Five entities cover v1; every record has `id`, `user_id`, `created_at` and `updated_at`.

| Entity | Key fields | Notes |
| --- | --- | --- |
| Goal | title, why, target\_date, success\_metric, cadence, stall\_threshold\_days, status (active, paused, done, dropped) | The "why" feeds every nudge |
| Task | goal\_id?, title, notes, status (todo, doing, done, skipped), priority (1–3), due\_date, estimate\_min, completed\_at, source (user, agent, mcp) | Agent-created tasks start as proposals until approved |
| CheckIn | date, kind (daily, stall, ad hoc), done\_summary, blockers, energy (1–5), agent\_reply | One row per conversation with the agent |
| Review | week\_start, per-goal progress, wins, slips, plan\_next\_week | Generated weekly, editable |
| AgentAction | actor (agent, mcp\_client, user), tool, input, result, reverted\_at | Audit log; every change is reversible |

Settings live in one `UserSettings` record (see Configuration).

## Functional requirements

Each requirement has an ID so tasks and tests can reference it.

**Goals and tasks**

- **FR-1** Create, edit, pause, complete and drop goals from the web app and via MCP.
- **FR-2** Create, edit, reorder and complete tasks, optionally linked to a goal.
- **FR-3** Show a "Today" view: top 1–3 tasks chosen by the agent plus anything due.
- **FR-4** Show progress per goal: tasks done per week against the goal's cadence.

**Agent**

- **FR-5** Break a goal into 3–7 concrete next actions; user approves, edits or rejects each.
- **FR-6** Daily check-in at the configured time: ask what got done, update tasks from the reply, propose today's top tasks.
- **FR-7** Stall detection: flag an active goal with no completed task in `stall_threshold_days`, ask what's blocking, offer to shrink the next step.
- **FR-8** Weekly review on the configured day: progress per goal, wins, slips, proposed plan for next week.
- **FR-9** Free-form chat with the agent in the web app, with access to the same tools as MCP clients.

**Notifications**

- **FR-10** Web push to the installed PWA for check-ins, stall alerts and reviews; tapping opens the conversation.
- **FR-11** Email fallback when push is disabled or unanswered after a configurable delay.
- **FR-12** Respect quiet hours and a daily nudge cap.

**Control and history**

- **FR-13** Every change made by the agent or an MCP client is logged and can be undone from the web app.
- **FR-14** Autonomy levels: suggest only, or allowed to create and reschedule tasks on its own.

## Agent behavior

The agent runs on a schedule, not an always-on loop: each job builds context, calls the LLM with tools, applies allowed changes, and sends one message.

**Context per run**

- Active goals with their "why", cadence and last progress date.
- Open tasks, overdue items, last 7 days of check-ins and the latest review.
- User settings: tone, autonomy level, quiet hours.

**Rules**

1. One ask per message: a single question or a single suggested next step.
2. Always tie a nudge to the goal's stated "why", never a generic pep talk.
3. When a goal stalls twice in a row, propose a smaller next step before pushing harder.
4. Never delete data; propose drops, the user confirms.
5. Under "suggest only", every write becomes a proposal the user approves.
6. Respect the daily nudge cap; stall alerts yield to the daily check-in if both fall on the same day.

**Tone presets** (system-prompt fragments)

- *Coach*: warm, celebrates progress, asks open questions.
- *Direct*: short, names the slip plainly, asks for a commitment with a time.
- *Drill sergeant*: blunt accountability, still respectful, never insulting.

**LLM access** goes through a Go `LLMProvider` interface (Claude first; OpenAI or a local model later), so the agent is model-agnostic.

## MCP server

The MCP server is a Go service using the official MCP Go SDK over Streamable HTTP, protected by OAuth 2.1, so remote clients (Claude, ChatGPT, Cursor, VS Code) can connect with a URL.

| Tool | Input | Returns |
| --- | --- | --- |
| `list_goals` | status? | goals with progress summary |
| `create_goal` | title, why, target\_date?, cadence? | goal |
| `update_goal` | id, fields | goal |
| `list_tasks` | goal\_id?, status?, due\_before? | tasks |
| `create_task` | title, goal\_id?, due\_date?, priority? | task |
| `update_task` | id, fields | task |
| `complete_task` | id, note? | task |
| `get_today` | — | top tasks + due items |
| `log_checkin` | done\_summary, blockers?, energy? | check-in |
| `get_progress` | goal\_id, period (week, month) | counts and trend |
| `get_stalled_goals` | — | goals past threshold |
| `propose_breakdown` | goal\_id | draft tasks for approval |

**Resources**: `goal://{id}` and `review://latest`, so clients can pull context without a tool call.

**Rules**: destructive actions are not exposed; every call is written to `AgentAction` with `actor = mcp_client` and the client name; rate limit of 60 calls per minute per token.

## Configuration

All behavior is driven by `UserSettings`, editable in the web app and read by the scheduler and the prompt builder on every run.

| Setting | Type | Default |
| --- | --- | --- |
| Tone | coach, direct, drill sergeant | direct |
| Autonomy | suggest only, act and log | suggest only |
| Daily check-in time | time | 08:30 |
| Weekly review day and time | weekday + time | Sunday 18:00 |
| Quiet hours | time range | 21:00–07:30 |
| Daily nudge cap | integer | 3 |
| Default stall threshold | days (per-goal override) | 3 |
| Channels | push, email | push |
| Email fallback delay | minutes | 120 |
| LLM provider and model | enum + string | Claude, latest Sonnet |
| LLM API key | encrypted secret | — |
| Language | en, es | en |
| Time zone | IANA zone | America/Santo\_Domingo |

## Architecture and tech stack

One Go binary serves the REST API, the MCP server and the scheduled agent, all sharing the same domain layer and database.

&#91;embedded content: system architecture · 3 clients, 1 Go backend, 3 external services\]

The PWA and MCP clients only pull; the agent is the one part that initiates, calling the LLM and sending push or email that opens the PWA.

| Layer | Choice |
| --- | --- |
| Backend | Go, chi router, sqlc for queries, goose migrations |
| MCP | Official MCP Go SDK, Streamable HTTP, OAuth 2.1 |
| Scheduler | In-process cron (robfig/cron) with a Postgres job lock |
| Database | PostgreSQL |
| Web app | Angular (standalone components, signals), Angular service worker for PWA and push |
| API contract | OpenAPI 3.1, generated TS client (Dart client in v2) |
| Auth | Email magic link for the app; OAuth 2.1 for MCP clients |
| Hosting | Docker; Fly.io or a small VPS; managed Postgres |
| CI | GitHub Actions: lint, tests, evals, deploy |

## Non-functional requirements

- **NFR-1 Security**: API keys encrypted at rest; OAuth tokens scoped per client and revocable from settings.
- **NFR-2 Privacy**: no task data sent to the LLM beyond what the current run needs; no training opt-in.
- **NFR-3 Reliability**: scheduled jobs are idempotent and run once even with two instances (Postgres advisory lock).
- **NFR-4 Latency**: REST and MCP reads under 200 ms p95; agent messages delivered within 2 minutes of schedule.
- **NFR-5 Cost**: under US$10/month for one user, including hosting and LLM calls.
- **NFR-6 Observability**: structured logs, a trace per agent run (prompt, tool calls, tokens, cost).
- **NFR-7 Offline**: the PWA shows cached Today and tasks offline and syncs edits when back online.
- **NFR-8 i18n**: English and Spanish UI and agent messages.

## Evals and testing

Agent quality is tested like code: fixed scenarios, expected behavior, run in CI on every prompt or model change.

**Agent evals** (seeded database + fake clock, graded by rules and an LLM judge)

- Picks sensible top tasks: overdue and high-priority first, respects goal cadence.
- Flags exactly the goals past their stall threshold, none early.
- Under "suggest only", makes zero direct writes.
- Nudge references the goal's "why" and asks one thing.
- Tone matches the configured preset.
- Parses a free-text check-in reply into the right task updates.

**Code tests**

- Go unit tests for domain services; integration tests against Postgres (testcontainers).
- MCP contract tests: each tool's schema and behavior, run through a real MCP client.
- Angular component tests and one Playwright end-to-end flow (create goal → check-in → complete task).

## Milestones and open questions

v1 fits in about 4 weeks of part-time work, each milestone usable on its own.

1. **Week 1 — Core**: Go project, Postgres schema, REST CRUD for goals and tasks, OpenAPI spec.
2. **Week 2 — MCP**: MCP server with all v1 tools and OAuth; used daily from Claude.
3. **Week 3 — Agent**: scheduler, daily check-in, stall detection, weekly review, first eval suite.
4. **Week 4 — PWA**: Angular app with Today, goals, chat, settings, push notifications; deploy.

The first real goal loaded into the system is the job search, so the demo runs on live data.

**Open questions**

- [ ] Product name for the repo and demo.
- [ ] Public demo with a read-only sample account, or private with a recorded walkthrough?
- [ ] Magic link or Google sign-in for the web app?
- [ ] Should check-in replies also be possible by answering the push notification inline?
