# AGENTS.md

Go HTTP API template (project bank went blank — the copy is the point). Starter scaffold with Postgres, migrations, lint and mocks already wired. Rename the `go.mod` module and create the domain module in `internal/` on first use.

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
handler → service → repository (port) → repository/postgres
                   → client (port)   → client/console
                   → model
```

Wire it in `<domain>/<domain>.go`; expose read adapters to other domains in `<domain>/public.go`. Cross-cutting concerns go in `internal/shared` (`core`, `customerr`, `httpx`); reusable infra in `pkg`.

Adding a feature follows the layer order:

1. **`model`** — constructor (`New<Model>(input)`), invariants, state changes as explicit methods using `core.Model.Touch`. IDs from `core.NewModel()`, soft delete via `DeletedAt`, money in integer cents.
2. **`repository`** — add the method to the port interface.
3. **`repository/postgres`** — prepared statements via `postgrex.Querier`; not-found returns `(nil, nil)`, never `sql.ErrNoRows`; callers check for `nil`.
4. **`service`** — use case wrapping infra failures as `customerr.Integration(err)` and domain outcomes as `NotFound`/`Conflict`/`Validation`. Multi-repo writes go through `core.TransactionManager[repository.Transaction]` (`postgrex.RunInTx`).
5. **`handler`** — parse via `httpx.ParseRequest`, call the service, answer with `httpx.Write*Response` or `httpx.WriteError` (error type maps to HTTP status).
6. **`http/route.go`** — register under `/api/v1/...`.

### Cross-module reads

A module never imports another module's service: to read another module's data, the consumer depends on a **client port** owned by the producer. The producer owns three pieces for each exposed read:

- **`<domain>/client/`** — the port: a read interface the consumer codes against (e.g. `PaymentReader.GetPayment`). Also the home of outbound ports the module itself consumes (e.g. `NotificationClient`, with its adapter in `client/console`).
- **`<domain>/public/`** — the adapter: a thin struct over a `service` that forwards read calls into the port shape (e.g. `public.NewPaymentReader(paymentService)`, whose `GetPayment` delegates to `paymentService.GetPayment`).
- **`<domain>/public.go`** — the assembly: a module-level constructor (e.g. `NewPaymentReader(db)`) that builds repositories + service and returns the adapter typed as the `client` port, so the consumer constructs it in a single call.

The consumer codes against the produced `client` port only; mocks are generated from it into `test/<domain>/mocks`.

## Rules

**Always:**

- Changes stay inside the active module
- Follow existing patterns and naming
- Wire features declaratively in `<module>.go`

**Never:**

- Add architectural layers
- Bypass use cases (handler → provider)
- Cross-import between modules
- Use globals or `panic()`
- Refactor unrelated code

## When in Doubt

1. Check `internal/note/` as reference
2. Pick the simplest change that keeps modules independent

## Tests

- External packages only (`handler_test`, `service_test`), table-driven, testify.
- Mocks generated with mockery into `test/<domain>/mocks` and `test/shared/mocks`; regenerate with `make mock` and commit them.
