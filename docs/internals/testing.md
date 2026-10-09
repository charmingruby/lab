# Testing

## Philosophy

- **No mocks.** Tests use real implementations. A test that needs a mock is testing the wrong seam.
- **Usecase tests prove business logic** — outcomes, state transitions, and transaction boundaries through the real stack (`usecase → repository + dependent clients`).
- **Delivery tests prove the edge** — how that logic behaves behind a transport (`http`, `grpc`, `queue`, ...): parsing, contract, and side-effects.
- **Model tests stay pure unit** — no infrastructure.

## Prerequisites

Each case needs a fresh, isolated, real golden source.

- Current: `test/container/postgres.go` spins one database per case and runs `db/migration`. Needs a reachable Docker daemon (`DOCKER_HOST` or `docker.host` in `~/.testcontainers.properties` for non-default sockets). Without Docker, point `TEST_DATABASE_URL` at a real database (`docker compose up -d` + `task mig-up`) and tests reuse it.

## Test conventions

- Table-driven, always. One behavior per row.
- Fresh stack per case. Rows that need prior state prepare it through the real stack; never share state between rows.
- Usecase rows assert the business outcome (result or typed error).
- Delivery rows assert the transport contract plus the persisted side-effect — never trust the response alone.
- External integrations use an in-memory fake with real state. The fake may inject errors for paths that cannot happen otherwise (e.g. "notification failure does not fail the usecase").
- Infra failures that cannot happen through the real stack are dropped, not faked.
- A test that needs `time.Sleep` to pass is wrong — it hides a race.
- Ship focused tests for the behavior you changed.

## Verifying

- Smallest proof that the change works: run the tests you touched with `go test ./internal/<domain>/... -run TestName -race` (needs a real golden source — container or `TEST_DATABASE_URL`).
- Targeted lint: `task lint` on the scope you changed. Do not run repo-wide lint unless asked.
- Backend behavior changes ship with focused tests for that behavior.
