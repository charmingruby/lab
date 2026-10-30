# Testing

## Philosophy

- **No mocks.** Tests use real implementations. A test that needs a mock is testing the wrong seam.
- **Usecase tests prove business logic** — outcomes, state transitions, and transaction boundaries through the real stack (`usecase → repository + dependent clients`).
- **Delivery tests prove the edge** — how that logic behaves behind a transport (`http`, `grpc`, `queue`, ...): parsing, contract, and side-effects.
- **Model tests stay pure unit** — no infrastructure.

## Prerequisites

Each case that uses persistence needs a fresh, isolated, real golden source. Model-only tests need no infrastructure. See `test/container/` and `db/migration` for the persistence setup.

## Test conventions

- Table-driven, always. One behavior per row.
- Fresh stack per case. Rows that need prior state prepare it through the real stack; never share state between rows.
- Usecase rows assert the business outcome (result or typed error).
- Delivery rows assert the transport contract plus the persisted side-effect — never trust the response alone.
- External integrations use a real local/sandboxed provider or an in-memory fake with real state, as described in [external-integrations.md](./external-integrations.md#testing-an-integration). The fake may inject errors for paths that cannot happen otherwise (e.g. "notification failure does not fail the usecase").
- Infra failures that cannot happen through the real stack are dropped, not faked.
- A test that needs `time.Sleep` to pass is wrong — it hides a race.
- Ship focused tests for the behavior you changed.

## Required coverage and placement

- Every usecase method requires focused tests for its business outcomes and relevant failure paths.
- Every inbound entry point requires its own tests: HTTP endpoint methods, listener `Handle` methods, gRPC handlers, and equivalent delivery methods. A thin delegation layer is not exempt; usecase coverage does not replace delivery coverage, and delivery coverage does not replace usecase coverage.
- Endpoint tests cover parsing/validation where applicable, response/error mapping, and observable side effects. A publishing endpoint must prove publication; a successful HTTP response alone is insufficient.
- Listener tests cover event-to-usecase mapping, persisted effects, failure propagation or intentional skips, and idempotency under repeated delivery. Broker-specific decoding and ack/redelivery are additionally tested at the adapter boundary.
- Keep each test beside the implementation it exercises: `usecase/<action>_test.go`, `http/endpoint/<action>_v<n>_test.go`, and `listener/on_<event>_v<n>/listener_test.go`. When the module has multiple inbound mechanisms, prepend `delivery/` to the transport paths, without adding `queue/`.
- Do not place tests at the module root or test the composition file (`<domain>/<domain>.go`) directly. For example, `internal/pingpong/pingpong_test.go` is not an allowed location. Tests of behavior belong in the responsible usecase or delivery package, even when they exercise the real stack. Test setup helpers stay with those tests, not in a root-level composition test.

## Verifying

- Smallest proof that the change works: run the tests you touched with `go test ./internal/<domain>/... -run TestName -race` (needs a real golden source — container or `TEST_DATABASE_URL`).
- Lint: run `task lint`. It formats Go files and lints the entire repository; inspect any resulting changes.
- Backend behavior changes ship with focused tests for that behavior.
