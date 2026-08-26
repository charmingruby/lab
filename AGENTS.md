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

**Delivery mechanism layout — a mechanism is any transport the domain speaks:**

- **One mechanism** (default, e.g. HTTP only): flat — `internal/<domain>/http/` holds `endpoint/` and `route.go`. No `delivery/` wrapper.
- **Two or more mechanisms**: nested — introduce `delivery/` as the parent. Move the existing `http/` to `delivery/http/`. Add the new mechanism beside it: `delivery/grpc/`, `delivery/queue/`.

Everything bound to a transport (DTOs, protos, endpoints, listeners, event schemas) lives in its own mechanism folder and is wired in `<domain>/<domain>.go`.

**Messaging is always a delivery mechanism.** A queue is a transport in both directions — a consumer is inbound (listens, like HTTP receives requests), a producer is outbound (delivers events to another system). Both live in `delivery/queue/` and both count toward the `delivery/` split: adding any queue to a domain that has HTTP forces the nested layout. Messaging never becomes a `client` integration.

**`repository/` shape** — template for any port with multiple backends: interface at the root (`repository/repository.go`), each implementation in its own subpackage (`repository/postgres/`). Same shape for `client/` (port in `client/notifier.go`, adapter in `client/console/`) and for messaging adapters inside `delivery/queue/` (`queue/kafka`, `queue/sqs`).

**External dependencies (storage, email, cache, third-party APIs)** — same shape as `client/`, raw connection in `pkg/`, port scoped to who consumes it (domain-specific vs. shared). Messaging is excluded — it's a delivery mechanism, not an integration. See [docs/external-integrations.md](docs/external-integrations.md).

Wire the module in `<domain>/<domain>.go`. Expose read adapters to other domains in `<domain>/public.go`. Cross-cutting concerns go in `internal/shared` (`core`, `customerr`, `httpx`). Reusable infra goes in `pkg`.

### Adding a feature — fixed layer order

1. **`model`** — constructor (`New<Model>(input)`), invariants, state changes as explicit methods using `core.Model.Touch`.
2. **`repository`** — add the method to the port interface.
3. **`repository/postgres`** — prepared statements via `postgrex.Querier`. Not-found returns `(nil, nil)`, never `sql.ErrNoRows`. Callers check for `nil`.
4. **`usecase`** — one file per implemented use case; all use case interfaces (the mocks mockery generates) are aggregated in a single `usecase.go`. Infra failures wrap as `customerr.Integration(err)`. Domain outcomes map to `NotFound`/`Conflict`/`Validation`. Multi-repo writes go through `core.TransactionManager[repository.Transaction]` (`postgrex.RunInTx`).
5. **`http/endpoint`** — parse via `httpx.ParseRequest`, call the use case, answer with `httpx.Write*Response` or `httpx.WriteError` (error type maps to HTTP status).
6. **`http/route.go`** — register under `/v1/...` (lands on `/api/v1/...`).

## Tests

- External packages only (`endpoint_test`, `usecase_test`), table-driven, testify.
- Mocks generated with mockery into `test/<domain>/mocks` and `test/shared/mocks`; regenerate with `make mock` and commit them.

## Progressive Documentation Loading

CRITICAL: Only load the reference below that matches your current task. Do NOT read all docs up front — most work needs nothing beyond this file.

| Task                                                                                                                    | Load                                                                            |
| ----------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| Implement a domain feature (model → repository → usecase → endpoint)                                                    | [docs/coding-patterns.md](docs/coding-patterns.md) + mirror `internal/billing/` |
| Add a layer, protocol, or module boundary                                                                               | "Structure" above + mirror `internal/billing/`                                  |
| Add or consume an external integration (storage, email, cache, third-party API)                                         | [docs/external-integrations.md](docs/external-integrations.md)                  |
| Expose an application boundary — new/handled protocol, or a versioned HTTP/gRPC/queue interface (incl. break/deprecate) | [docs/versioning.md](docs/versioning.md)                                        |
| Read another module's data                                                                                              | [docs/cross-module-reads.md](docs/cross-module-reads.md)                        |
| Tests                                                                                                                   | "Tests" above + existing `*_test.go` in the domain                              |

## Rules

**Always:**

- Keep changes inside the active module.
- Match naming and layout in `internal/billing/` exactly — new domain, same shape.
- Wire features declaratively in `<module>.go`.

**Never:**

- Add a layer not listed in Structure (no extra abstraction between usecase and repository).
- Skip the usecase (handler/endpoint calling repository or client directly).
- Import another module's package (`usecase`, `repository`, `model`) — use its `client` port instead.
- Use globals or `panic()`.
- Touch files outside the current task's domain.

## When in Doubt

1. Copy the pattern from `internal/billing/`.
2. If `internal/billing/` doesn't cover the case, pick the option that adds the least new structure.
