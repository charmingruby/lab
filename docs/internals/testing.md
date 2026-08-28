# Testing

## Test conventions

- External packages only (`endpoint_test`, `usecase_test`), table-driven, testify.
- Mocks generated with mockery into `test/<domain>/mocks` and `test/shared/mocks`; regenerate with `task mock` and commit them.
- A test that needs a `time.Sleep` to pass is wrong — it is hiding a race condition or missing synchronization.
- Ship focused tests for the behavior you changed. Do not run repo-wide checks unless the developer asks.

## Verifying

- Smallest proof that the change works: run the tests you touched with `go test ./internal/<domain>/... -run TestName -race`.
- Targeted lint: `task lint` on the scope you changed. Do not run repo-wide lint unless asked.
- Backend behavior changes ship with focused tests for that behavior.
- Upon request, integration-test against real Postgres with `docker compose up -d` and `task test`.
