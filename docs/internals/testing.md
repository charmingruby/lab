# Testing

## Philosophy

- **No mocks.** Tests use real implementations: the real postgres for the
  golden source, real usecases behind HTTP endpoints, and in-memory fakes for
  external clients. A test that needs a mock is testing the wrong seam.
- **Usecase tests validate the logic itself** — business outcomes against a
  real database (`invalid → Validation`, `missing → NotFound`, state
  transitions, transaction boundaries).
- **HTTP tests validate the edge integration** — request parsing, status
  codes, response shape, and persistence side-effects through the full
  `endpoint → usecase → repository` stack.
- **Model tests stay pure unit** — constructors and state transitions need no
  infrastructure.

## Prerequisites

Integration tests need a reachable Docker daemon (testcontainers spins
postgres per test). Default installs work out of the box; non-default
sockets (OrbStack, rootless Docker, Colima) need one of:

- `DOCKER_HOST` pointing at the daemon socket, or
- `docker.host` in `~/.testcontainers.properties`, e.g.
  `docker.host=unix:///Users/you/.orbstack/run/docker.sock`.

No Docker at all? Point `TEST_DATABASE_URL` at a real postgres
(`docker compose up -d` + `task mig-up`) and tests reuse it instead of
containers.

## Test conventions

- External packages only (`endpoint_test`, `usecase_test`), testify for asserts.
- **Table-driven, always.** Every test is a `tests := []struct{...}` slice plus
  a `for _, tt := range tests { t.Run(tt.name, ...) }` loop. One behavior per
  row: `name`, input, `wantErr`/`errType` (or `wantStatus` on the edge), and
  the expected outcome. Table first, loop second — never ad-hoc subtests.
- **Seed per case, isolate per case.** Rows that need prior state carry a
  `seed` (a bool for "create one" or a func that prepares state through the
  real stack and returns what the case needs). The stack setup runs inside
  each subtest, so every row gets a fresh database.
- Golden source is always real postgres: `container.StartPostgres(t)` from
  `test/container/postgres.go` spins the container, runs `db/migration`, and
  cleans up. One fresh database per test case.
- HTTP rows assert the edge contract (`wantStatus` + body check) and, when
  the case mutates state, verify the side-effect afterwards (`wantPersisted`,
  `wantAssigned`) through the usecase — never trust the response alone.
- External clients use the `client/memory/` fake (real state, no
  expectations). The fake can inject errors via `SetError` for paths like
  "notification failure does not fail the usecase". Prefer the real
  integration when viable; `memory/` is the trivial fallback.
- Infra failures that cannot happen through the real stack (e.g. "db
  connection refused") are not faked — drop those cases instead of mocking them.
- A test that needs a `time.Sleep` to pass is wrong — it is hiding a race condition or missing synchronization.
- Ship focused tests for the behavior you changed. Do not run repo-wide checks unless the developer asks.

## Verifying

- Smallest proof that the change works: run the tests you touched with `go test ./internal/<domain>/... -run TestName -race` (needs docker for postgres containers, or `TEST_DATABASE_URL` pointing at a real postgres).
- Targeted lint: `task lint` on the scope you changed. Do not run repo-wide lint unless asked.
- Backend behavior changes ship with focused tests for that behavior.
