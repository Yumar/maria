# T-005: OpenAPI foundation, codegen and router wiring

- **Story**: S-003
- **FR/NFR**: FR-1 (groundwork), NFR-4
- **Complexity**: senior
- **Depends on**: T-004
- **Establishes pattern**: yes (spec-first API, strict handlers, problem+json errors)

## Goal
Set up the spec-first API: a hand-written OpenAPI 3.1 document, oapi-codegen generating chi strict-server types, request validation against the spec, RFC 9457 error responses and dev auth. Prove it all end to end with `GET /api/v1/me`.

## Files to touch
- `api/openapi.yaml` — create:
  - `openapi: 3.1.0`, `info`, `servers: [{url: /api/v1}]`
  - `components.schemas`:
    - `Problem`: RFC 9457 `type`, `title`, `status`, `detail`, plus `errors: [{field, message}]`
    - `User`: `id` (uuid), `email`
  - `components.responses`: `BadRequest`, `Unauthorized`, `NotFound`, `ValidationFailed`, all with `application/problem+json`
  - path `GET /me` → 200 `User`, 401
  - `/healthz` is **not** in the spec
  - Nullable fields use the 3.1 style `type: [T, "null"]`
- `backend/oapi-codegen.yaml` — create:
  - package `gen`
  - generate `chi-server`, `strict-server`, `models`, `embedded-spec`
  - output `internal/httpapi/gen/api.gen.go`
  - `output-options.nullable-type: true` (PATCH bodies must tell "absent" apart from `null`)
- `backend/internal/httpapi/gen/api.gen.go` — generated, committed.
- `backend/internal/httpapi/server.go` — create:
  - `type Server struct{ queries *sqlc.Queries }` plus a constructor
  - `var _ gen.StrictServerInterface = (*Server)(nil)`
  - `GetMe` reads the user ID from `auth.UserIDFrom` and loads it with `GetUserByID`
  - Later tasks add service fields to `Server`, each typed as a small interface declared in `httpapi`
- `backend/internal/httpapi/errors.go` — create:
  - `writeProblem(w, status, title, detail string, fields []apperr.FieldError)`
  - `problemFromError(err) (status int, …)`, mapping `apperr` kinds to statuses: NotFound → 404, Validation → 422, Conflict → 409, Unauthorized → 401, anything else → 500. For 500s, log the error and send the generic detail "internal error" to the client.
  - strict-handler `RequestErrorHandlerFunc` (bad or undecodable body → 400) and `ResponseErrorHandlerFunc` (→ `problemFromError`)
- `backend/internal/httpapi/errors_test.go` — create: the mapping table and that 500s don't leak details.
- `backend/internal/httpapi/router.go` — modify: `NewRouter(logger, deps)`. Mount `/api/v1` with:
  1. oapi-codegen `nethttp-middleware` `OapiRequestValidator` on the embedded spec. Its failures become problem+json: 400 for malformed input, 422 with `errors[]` for schema violations. Its server-URL prefix must be handled so `/api/v1` paths match.
  2. `auth.Middleware` with a problem+json `onError` (401).
  3. The strict handler.
  Unknown `/api/v1` routes return a problem+json 404.
- `backend/internal/httpapi/server_test.go` — create:
  - `/api/v1/me`: 200 with a stub authenticator + fake queries (or an integration-tagged test with testutil); 401 when the authenticator fails
  - unknown route 404 as problem+json
  - a request that violates the schema returns 422 (e.g. add a test-only path via a param to `/me`, or cover this in T-007; state which)
- `backend/cmd/server/main.go` — modify: `EnsureDevUser` at startup when `AUTH_MODE=dev` (log the dev user email at WARN level), build `DevAuthenticator` and pass the deps to `NewRouter`.
- `backend/go.mod` — modify: add tool `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` (**v2.8.0 or newer**, for OpenAPI 3.1), runtime `github.com/oapi-codegen/runtime` ≥ v1.6.0, `github.com/oapi-codegen/nethttp-middleware`, and the `kin-openapi` version they require.
- `Makefile` — modify: `generate` also runs `go tool oapi-codegen -config oapi-codegen.yaml ../api/openapi.yaml`.
- `backend/.golangci.yml` — modify if needed so `internal/httpapi/gen` is excluded.

## Reference (copy this pattern)
- `backend/internal/httpapi/router.go` (T-001), `backend/internal/auth/auth.go` and `backend/internal/apperr/apperr.go` (T-004).

## Acceptance criteria
- [ ] `api/openapi.yaml` declares `openapi: 3.1.0` and generates without warnings; `make check-generated` passes.
- [ ] `curl localhost:8080/api/v1/me` returns the dev user as JSON.
- [ ] Every error response from `/api/v1` has `Content-Type: application/problem+json` and matches the `Problem` schema.
- [ ] A 500 never exposes internal error text to the client, but it is logged with the request ID.
- [ ] Tests listed above pass, alongside `make lint`, `make test` and `make test-integration`.
- [ ] Report back briefly on how `nullable-type` renders PATCH fields, so T-007 can follow it.

## Verify with
```bash
make generate && make check-generated
make lint
make test
make test-integration
make db-up && make run &   # then:
curl -fsS localhost:8080/api/v1/me
curl -s localhost:8080/api/v1/nope -i | head -5
```

## Out of scope
- Goal and task endpoints (T-007, T-008).
- TS client generation (Week 4).
- Security schemes for real auth (Week 4 / MCP OAuth in Week 2).
