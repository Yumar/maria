# T-004: Error model and dev auth stub

- **Story**: S-003
- **Issue**: [#9](https://github.com/Yumar/maria/issues/9)
- **FR/NFR**: FR-1 (groundwork), NFR-1
- **Complexity**: implementer
- **Depends on**: T-002
- **Establishes pattern**: yes (small, fully specified below)

## Goal
Add a domain error type that every service returns, and an authentication seam that puts the current user ID into the request context. In Week 1 the authenticator is a dev stub for one seeded user. Magic-link auth replaces it in Week 4 without touching handlers.

## Files to touch
- `backend/internal/apperr/apperr.go` — create:
  ```go
  type Kind string // KindNotFound, KindValidation, KindConflict, KindUnauthorized
  type FieldError struct { Field, Message string }
  type Error struct { Kind Kind; Message string; Fields []FieldError }
  func (e *Error) Error() string
  func NotFound(msg string) *Error
  func Validation(msg string, fields ...FieldError) *Error
  func Conflict(msg string) *Error
  func Unauthorized(msg string) *Error
  func KindOf(err error) (Kind, bool) // uses errors.As; works on wrapped errors
  ```
- `backend/internal/apperr/apperr_test.go` — create: constructors, `KindOf` on wrapped (`fmt.Errorf("%w")`) and non-apperr errors.
- `backend/internal/auth/auth.go` — create:
  ```go
  func WithUserID(ctx context.Context, id uuid.UUID) context.Context
  func UserIDFrom(ctx context.Context) (uuid.UUID, bool)
  type Authenticator interface { Authenticate(r *http.Request) (uuid.UUID, error) }
  func Middleware(a Authenticator, onError func(w http.ResponseWriter, r *http.Request, err error)) func(http.Handler) http.Handler
  type DevAuthenticator struct { UserID uuid.UUID } // always returns UserID
  func EnsureDevUser(ctx context.Context, q *sqlc.Queries, email string) (uuid.UUID, error) // UpsertUserByEmail
  ```
  `onError` lets the HTTP layer (T-005) render problem+json without `auth` importing `httpapi`.
- `backend/internal/auth/auth_test.go` — create: middleware sets the user ID on success, calls `onError` and does not call next on failure; `UserIDFrom` returns false on an empty ctx.
- `backend/internal/config/config.go` — modify: add `AUTH_MODE` (only `dev` is accepted for now, default `dev`) and `DEV_USER_EMAIL` (default `dev@localhost`). `Load` must **return an error when `APP_ENV=production` and `AUTH_MODE=dev`**.
- `backend/internal/config/config_test.go` — modify: cover the new vars and the production + dev rejection.
- `.env.example` — modify: add `AUTH_MODE`, `DEV_USER_EMAIL`.

## Reference (copy this pattern)
- `backend/internal/config/config.go` and `config_test.go` (T-001) — env parsing and table-driven test style.
- `backend/internal/db/queries/users.sql` / generated `UpsertUserByEmail` (T-002).

## Acceptance criteria
- [ ] APIs match the signatures above exactly (T-005 and T-007 build on them).
- [ ] `apperr` has no dependencies outside the stdlib; `auth` does not import `httpapi`.
- [ ] Production + dev auth mode fails config loading with a clear message.
- [ ] Unit tests cover all cases listed above.
- [ ] `make lint` and `make test` pass.

## Verify with
```bash
make lint
make test
```

## Out of scope
- Wiring into the router and `main.go` (T-005).
- HTTP status mapping and problem+json rendering (T-005).
- Real auth (magic link, Week 4) and MCP OAuth 2.1 (Week 2).
- Do not add new dependencies (uuid and pgx already exist).
