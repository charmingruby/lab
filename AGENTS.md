# AGENTS.md

Go backend template (project bank went blank — the copy is the point). Starter scaffold with Postgres, migrations, lint and mocks already wired. Rename the `go.mod` module and create the domain module in `internal/` on first use.

## Commands

- Local Infrastructure: `docker compose up -d` (matches `.env.example`).
- Dev server: `air` (builds `./cmd/api/main.go`, hot reload).
- Tests: `make test` — regenerates mocks first, so it needs `mockery` installed; then `go test ./... -race`.
- Lint: `make lint` (strict config in `.golangci.yml`; `make lint-fix` to auto-fix).
- Migrations: `make new-mig NAME=<name>` / `mig-up` / `mig-down` on `db/migration`.
- All other scripts live in the `Makefile`.

## Structure

The template ships one example domain as the blueprint; a domain is a **ports and adapters** module in `internal/<domain>`, one dependency spine:

```
<protocol> → usecase → repository (port) → repository/postgres
                              → client (port)   → client/console
                              → model
```

The structure **grows with the system** — it does not pre-ship empty folders. With a single inbound protocol, that protocol is the domain's main communication method, and the layer sits flat — e.g. `internal/<domain>/http/` with endpoints and routes for HTTP, no `delivery/` wrapper. The moment a second protocol is added (gRPC, a queue consumer), `delivery/` appears as the parent, the existing `http/` moves under `delivery/http/`, and the new one lands beside it in `delivery/grpc/` or `delivery/queue/`. Everything bound to a transport (DTOs, protos, endpoints, listeners) lives in its folder, and each is wired in `<domain>/<domain>.go`.

One spine is the `repository/` shape, the template for any port with multiple backends: the **interface lives at the root** (`repository/repository.go`), each implementation gets its own subpackage (`repository/postgres/`). A messaging port follows the same shape — interface and message types at the root, one subpackage per implementation (`queue/kafka`, `queue/sqs`) — just like `client/` (port in `client/notifier.go`, adapter in `client/console/`).

Wire the module in `<domain>/<domain>.go`; expose read adapters to other domains in `<domain>/public.go`. Cross-cutting concerns go in `internal/shared` (`core`, `customerr`, `httpx`); reusable infra in `pkg`.

Adding a feature follows the layer order:

1. **`model`** — constructor (`New<Model>(input)`), invariants, state changes as explicit methods using `core.Model.Touch`. IDs from `core.NewModel()`, soft delete via `DeletedAt`, money in integer cents.
2. **`repository`** — add the method to the port interface.
3. **`repository/postgres`** — prepared statements via `postgrex.Querier`; not-found returns `(nil, nil)`, never `sql.ErrNoRows`; callers check for `nil`.
4. **`usecase`** — one file per use case, each with its own port interface, wrapping infra failures as `customerr.Integration(err)` and domain outcomes as `NotFound`/`Conflict`/`Validation`. Multi-repo writes go through `core.TransactionManager[repository.Transaction]` (`postgrex.RunInTx`).
5. **`http/endpoint`** — parse via `httpx.ParseRequest`, call the use case, answer with `httpx.Write*Response` or `httpx.WriteError` (error type maps to HTTP status).
6. **`http/route.go`** — register under `/api/v1/...`.

### Versioning

Delivery is versioned, and HTTP is just the first application of it. In the flat layout, each endpoint file carries its own version suffix (`create_payment_v1.go`); a breaking change adds `create_payment_v2.go` beside it rather than mutating the file — only that endpoint versions, the rest of `endpoint/` stays on v1. `route.go` stays unversioned: it's the single registrar mapping each endpoint version to its route (`/api/v1/...`, `/api/v2/...` alongside it, deprecating `/v1` instead of mutating it). The same idea applies to any delivery form: a gRPC service registers `ServiceV1`/`ServiceV2`, and a message consumer binds versioned events or queues. Version what callers pin to (endpoint files, routes, services, message envelopes), never the inner layers — `model/`, `usecase/` and `repository/` stay version-free.

### Cross-module reads

A module never imports another module's use case: to read another module's data, the consumer depends on a **client port** owned by the producer. The producer owns three pieces for each exposed read:

- **`<domain>/client/`** — the port: a read interface the consumer codes against (e.g. `PaymentReader.GetPayment`). Also the home of outbound ports the module itself consumes (e.g. `NotificationClient`, with its adapter in `client/console`).
- **`<domain>/public/`** — the adapter: a thin struct over a use case that forwards read calls into the port shape (e.g. `public.NewPaymentReader(getPayment)`, whose `GetPayment` delegates to `getPayment.GetPayment`).
- **`<domain>/public.go`** — the assembly: a module-level constructor (e.g. `NewPaymentReader(db)`) that builds repositories + use case and returns the adapter typed as the `client` port, so the consumer constructs it in a single call.

The consumer codes against the produced `client` port only; mocks are generated from it into `test/<domain>/mocks`.

## Rules

**Always:**

- Changes stay inside the active module
- Follow existing patterns and naming
- Wire features declaratively in `<module>.go`

**Never:**

- Add architectural layers
- Bypass use cases (endpoint → provider)
- Cross-import between modules
- Use globals or `panic()`
- Refactor unrelated code

## When in Doubt

1. Check `internal/billing/` as reference
2. Pick the simplest change that keeps modules independent

## Tests

- External packages only (`endpoint_test`, `usecase_test`), table-driven, testify.
- Mocks generated with mockery into `test/<domain>/mocks` and `test/shared/mocks`; regenerate with `make mock` and commit them.
