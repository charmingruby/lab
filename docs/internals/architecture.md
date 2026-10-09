# Architecture

Ports and adapters. Every domain is a self-contained module with the same dependency spine.

## Request lifecycle

Inbound transport resolves to an endpoint, which follows the domain spine (see [Domain structure](#domain-structure)):

```
delivery → endpoint → usecase → model
                           → repository (concrete golden source)
                           → client port → adapter
```

- Current: chi mounted at `/api` with `httpx` helpers. Source: `cmd/api/main.go`, `internal/platform/httpx/`.

## Domain wiring

Each domain exposes one composition root in `<domain>.go` — it builds dependencies and registers routes. No wiring happens in endpoints or usecases.

- Current: `New(r chi.Router, db *sqlx.DB) error`, builds transaction manager, postgres repositories, usecases, endpoints.

Source: [internal/ticket/ticket.go](../../internal/ticket/ticket.go)

## Where code lives

- `cmd/api/main.go` — boots server, mounts domains, starts listening.
- `internal/<domain>/` — one ports-and-adapters module per bounded context.
  - `<domain>.go` — composition root (wires everything)
  - `public.go` — cross-domain read assembly
  - `http/endpoint/` — request DTO + handler (one delivery mechanism, see below)
  - `http/route.go` — registers versioned routes
  - `usecase/` — business actions, one file per use case
  - `model/` — entities, invariants, state changes
  - `repository/` — the golden source, concrete (no interface — swapping it is trivial when needed). Single-repo writes need no explicit transaction; the repository handles it.
  - `client/` — outbound ports this domain consumes (interfaces) plus adapters in subdirectories. Ports exist here as an anti-corruption layer for swappable providers. Never hosts exposed reads — those live in `public/`.
- `internal/shared/` — shared domain language: domain types, typed errors, and ports + adapters used by two or more domains. See the directory for the current set.
- `internal/platform/` — internal infrastructure with zero domain awareness: raw external clients and transport/config helpers. See the directory for the current set.
- `pkg/` — reserved for code exposed to the outside world.

- Current: postgres; `client/` adapters per provider. See the directories for the current set.

## Cross-domain communication

Domains never import each other's `usecase`, `repository`, or `model` packages. Instead:

1. The producing domain defines a read port in `public/` (not in `client/` — `client/` is outbound only).
2. A **public adapter** in `public/` wraps a usecase and implements the port.
3. **`public.go`** assembles the wiring and returns the adapter typed as the port.
4. The consuming domain depends only on the public port interface.

Source: [cross-module-reads.md](./cross-module-reads.md), [internal/ticket/public.go](../../internal/ticket/public.go)

## Error flow

Usecases map domain outcomes to typed errors; endpoints map typed errors to transport statuses. Usecases never return raw strings or unwrapped infrastructure errors.

- Current: `customerr.NotFound → 404`, `Conflict → 409`, `Validation → 422`, `Integration → 500`, via `httpx.WriteError`.

Source: [internal/shared/customerr/customerr.go](../../internal/shared/customerr/customerr.go)

## Transaction safety

Multi-repo writes run atomically through a transaction holding the repositories needed inside it. Single-repo writes do not need explicit transactions.

- Current: `repository.TransactionManager` + `repository.Transaction`, backed by `postgrex.RunInTx`.

Source: [internal/ticket/repository/transaction_manager.go](../../internal/ticket/repository/transaction_manager.go)

## Domain structure

A domain is a **ports and adapters** module in `internal/<domain>`. Canonical dependency spine (the only diagram — everything else links here):

```
endpoint → usecase → repository (concrete golden source)
                   → client (port) → adapter
                   → model
```

### Delivery mechanism layout

A mechanism is any transport the domain speaks:

- **One mechanism** (default, e.g. HTTP only): flat — `internal/<domain>/http/` holds `endpoint/` and `route.go`. No `delivery/` wrapper.
- **Two or more mechanisms**: nested — introduce `delivery/` as the parent. Move the existing `http/` to `delivery/http/`. Add the new mechanism beside it: `delivery/grpc/`, `delivery/queue/`.

Everything bound to a transport (DTOs, protos, endpoints, listeners, event schemas) lives in its own mechanism folder and is wired in `<domain>/<domain>.go`.

**Messaging is always a delivery mechanism.** A queue is a transport in both directions — a consumer is inbound (listens, like HTTP receives requests), a producer is outbound (delivers events to another system). Both live in `delivery/queue/` and both count toward the `delivery/` split: adding any queue to a domain that has HTTP forces the nested layout. Messaging never becomes a `client` integration.

### Repository and client shape

`repository/` is concrete — no port interface, no subpackages. `client/` keeps the port-adapter split (port + one subpackage per adapter). Same adapter-per-subpackage shape for messaging adapters inside `delivery/queue/`.

- Current: postgres; `client/` adapters per provider; queue adapters per provider. See the directories for the current set.

### External dependencies

Storage, email, cache, third-party APIs — same shape as `client/`, raw connection in `internal/platform/`, port scoped to who consumes it (domain-specific vs. shared). Messaging is excluded — it's a delivery mechanism, not an integration. See [external-integrations.md](./external-integrations.md).

Wire the module in `<domain>/<domain>.go`. Expose read adapters to other domains in `<domain>/public.go`. Shared domain language goes in `internal/shared/`. Connections and external config go in `internal/platform/`.

## Related

- [Glossary](./glossary.md) — domain terms with source file links
- [Coding patterns](./coding-patterns.md) — complete implementation reference per layer
- [External integrations](./external-integrations.md) — storage, email, cache, third-party APIs
- [Versioning](./versioning.md) — endpoint-level versioning
