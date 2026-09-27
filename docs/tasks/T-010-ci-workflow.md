# T-010: CI workflow (GitHub Actions)

- **Story**: S-001
- **Issue**: [#7](https://github.com/Yumar/maria/issues/7)
- **FR/NFR**: — (spec: "CI: GitHub Actions: lint, tests, evals, deploy")
- **Complexity**: implementer
- **Depends on**: T-005 (all `make` targets exist). Can run in parallel with T-006–T-009 because it touches only `.github/`.
- **Establishes pattern**: no

## Goal
Run lint, unit tests, integration tests and the generated-code drift check on every push to `main` and on every pull request.

## Files to touch
- `.github/workflows/ci.yml` — create:
  - Triggers: `push` (branches: `main`), `pull_request`. `concurrency` cancels in-progress runs for the same ref. `permissions: contents: read`.
  - Shared setup: `actions/checkout`, then `actions/setup-go` with `go-version-file: backend/go.mod` and the module cache (`cache-dependency-path: backend/go.sum`).
  - Jobs, all on `ubuntu-latest`:
    1. `lint`: `golangci/golangci-lint-action` with `working-directory: backend` and a pinned golangci-lint v2 version.
    2. `test`: `make test`.
    3. `integration`: `make test-integration` (Docker is available on GitHub-hosted runners for testcontainers).
    4. `generated`: `make check-generated`.
  - Pin every action to a released major version that is at least 7 days old (e.g. `actions/checkout@v5`). No floating `@main`.

## Reference (copy this pattern)
- `Makefile` targets (T-001, T-002, T-005). CI must call `make` targets, not duplicate their commands.

## Acceptance criteria
- [ ] Workflow YAML is valid (`actionlint` if available, otherwise explain how you checked it).
- [ ] Each job calls only `make` targets (or the lint action) and would fail on a non-zero exit.
- [ ] No secrets are required or referenced.

## Verify with
```bash
# if installed:
actionlint .github/workflows/ci.yml
# otherwise run the same targets locally:
make lint && make test && make test-integration && make check-generated
```

## Out of scope
- Eval job (Week 3), deploy job (Week 4), web/Angular jobs (Week 4).
- Branch protection settings.
