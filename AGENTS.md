# AGENTS.md

Go backend template. Starter scaffold with Postgres, migrations, lint and mocks already wired. Rename the `go.mod` module and create the domain module in `internal/` on first use.

## Commands

- Local Infrastructure: `docker compose up -d` (matches `.env.example`).
- Dev server: `air` (builds `./cmd/api/main.go`, hot reload).
- Tests: `make test` — regenerates mocks first, so it needs `mockery` installed; then `go test ./... -race`.
- Lint: `make lint` (strict config in `.golangci.yml`; `make lint-fix` to auto-fix).
- Migrations: `make new-mig NAME=<name>` / `mig-up` / `mig-down` on `db/migration`.
- All other scripts live in the `Makefile`.

## Structure

A domain is a **ports and adapters** module in `internal/<domain>`, one dependency spine:

```
<protocol> → usecase → repository (port) → repository/postgres
                              → client (port)   → client/console
                              → model
```

**Inbound protocol layout — pick one:**

- **One protocol** (default, e.g. HTTP only): flat — `internal/<domain>/http/` holds `endpoint/` and `route.go`. No `delivery/` wrapper.
- **Two or more protocols**: nested — introduce `delivery/` as the parent. Move the existing `http/` to `delivery/http/`. Add the new protocol beside it: `delivery/grpc/`, `delivery/queue/`.

Everything bound to a transport (DTOs, protos, endpoints, listeners) lives in its own protocol folder and is wired in `<domain>/<domain>.go`.

**`repository/` shape** — template for any port with multiple backends: interface at the root (`repository/repository.go`), each implementation in its own subpackage (`repository/postgres/`). Same shape for any messaging port (`queue/kafka`, `queue/sqs`) and for `client/` (port in `client/notifier.go`, adapter in `client/console/`).

Wire the module in `<domain>/<domain>.go`. Expose read adapters to other domains in `<domain>/public.go`. Cross-cutting concerns go in `internal/shared` (`core`, `customerr`, `httpx`). Reusable infra goes in `pkg`.

### Adding a feature — fixed layer order

1. **`model`** — constructor (`New<Model>(input)`), invariants, state changes as explicit methods using `core.Model.Touch`.
2. **`repository`** — add the method to the port interface.
3. **`repository/postgres`** — prepared statements via `postgrex.Querier`. Not-found returns `(nil, nil)`, never `sql.ErrNoRows`. Callers check for `nil`.
4. **`usecase`** — one file per implemented use case; all use case interfaces (the mocks mockery generates) are aggregated in a single `usecase.go`. Infra failures wrap as `customerr.Integration(err)`. Domain outcomes map to `NotFound`/`Conflict`/`Validation`. Multi-repo writes go through `core.TransactionManager[repository.Transaction]` (`postgrex.RunInTx`).
5. **`http/endpoint`** — parse via `httpx.ParseRequest`, call the use case, answer with `httpx.Write*Response` or `httpx.WriteError` (error type maps to HTTP status).
6. **`http/route.go`** — register under `/api/v1/...`.

### Versioning

Breaking or deprecating an API version, or versioning a gRPC/queue boundary? Read [docs/versioning.md](docs/versioning.md).

### Cross-module reads

Reading another module's data? Read [docs/cross-module-reads.md](docs/cross-module-reads.md).

## Rules

**Always:**

- Keep changes inside the active module.
- Match naming and layout in `internal/billing/` exactly — new domain, same shape.
- Wire features declaratively in `<module>.go`.

**Never:**

- Add a layer not listed in Structure (no service layer, no extra abstraction between usecase and repository).
- Skip the usecase (handler/endpoint calling repository or client directly).
- Import another module's package (`usecase`, `repository`, `model`) — use its `client` port instead.
- Use globals or `panic()`.
- Touch files outside the current task's domain.

## When in Doubt

1. Copy the pattern from `internal/billing/`.
2. If `internal/billing/` doesn't cover the case, pick the option that adds the least new structure.

## Tests

- External packages only (`endpoint_test`, `usecase_test`), table-driven, testify.
- Mocks generated with mockery into `test/<domain>/mocks` and `test/shared/mocks`; regenerate with `make mock` and commit them.
