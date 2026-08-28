# Architecture

The Go backend template follows a **ports and adapters** architecture. Every domain is a self-contained module with the same dependency spine. This document describes the runtime flow and how the pieces connect.

## Request lifecycle

```
cmd/api/main.go
  → chi router (mounted at /api)
    → domain middleware (each <domain>.go wires itself)
      → versioned route (/v1/<resources>)
        → endpoint (parses request, calls usecase)
          → usecase (business logic, transactions)
            → model (invariants, state changes)
            → repository/port (interface)
              → repository/postgres (prepared statements)
            → client/port (interface, if cross-domain or external)
              → client/<adapter>/ (concrete implementation)
          → httpx.Write*Response → HTTP response
```

## Domain wiring

Each domain exposes one `New(r chi.Router, db *sqlx.DB) error` function in `<domain>.go`. This is the composition root — it builds the transaction manager, postgres repositories, usecases, endpoints, and registers routes. No wiring happens in endpoints or usecases.

Source: [internal/ticket/ticket.go](../../internal/ticket/ticket.go)

## Where code lives

- `cmd/api/main.go` — boots server, mounts domains, starts listening.
- `internal/<domain>/` — one ports-and-adapters module per bounded context.
  - `<domain>.go` — composition root (wires everything)
  - `public.go` — cross-domain read assembly
  - `http/endpoint/` — request DTO + handler
  - `http/route.go` — registers versioned routes
  - `usecase/` — business actions, one file per use case
  - `model/` — entities, invariants, state changes
  - `repository/` — persistence port (interface), implementations in `postgres/`
  - `client/` — outbound ports (interfaces), adapters in subdirectories
- `internal/shared/` — cross-cutting: `core` (model base, pagination, transactions), `customerr` (typed errors), `httpx` (HTTP utils).
- `pkg/` — reusable infrastructure wrappers (raw SDK clients, DB drivers). Zero domain awareness.

## Cross-domain communication

Domains never import each other's `usecase`, `repository`, or `model` packages. Instead:

1. The producing domain defines a **client port** interface in `client/`.
2. A **public adapter** wraps a usecase and implements the port in `public/`.
3. **`public.go`** assembles the wiring and returns the adapter typed as the port.
4. The consuming domain depends only on the client port interface.

Source: [cross-module-reads.md](./cross-module-reads.md), [internal/ticket/public.go](../../internal/ticket/public.go)

## Error flow

Usecases map domain outcomes to typed errors:

| Outcome | Error type | HTTP status |
|---------|-----------|-------------|
| Resource not found | `customerr.NotFound` | 404 |
| Duplicate / conflict | `customerr.Conflict` | 409 |
| Invalid input | `customerr.Validation` | 422 |
| Infrastructure failure | `customerr.Integration` | 500 |

Endpoints pass errors to `httpx.WriteError`, which maps the `customerr` type to the correct HTTP status. Usecases never return raw strings or `fmt.Errorf` without a `customerr` wrapper.

Source: [internal/shared/customerr/customerr.go](../../internal/shared/customerr/customerr.go)

## Transaction safety

Multi-repo writes (e.g., updating two repositories atomically) go through `core.TransactionManager[repository.Transaction]`. The transaction struct holds the repo interfaces needed inside the transaction. Single-repo writes do not need explicit transactions — the postgres adapter handles it.

Source: [internal/shared/core/transaction.go](../../internal/shared/core/transaction.go), [internal/ticket/repository/postgres/transaction_manager.go](../../internal/ticket/repository/postgres/transaction_manager.go)

## Domain structure

A domain is a **ports and adapters** module in `internal/<domain>`, one dependency spine:

```
<protocol> → usecase → repository (port) → repository/postgres
                              → client (port)   → client/console
                              → model
```

### Delivery mechanism layout

A mechanism is any transport the domain speaks:

- **One mechanism** (default, e.g. HTTP only): flat — `internal/<domain>/http/` holds `endpoint/` and `route.go`. No `delivery/` wrapper.
- **Two or more mechanisms**: nested — introduce `delivery/` as the parent. Move the existing `http/` to `delivery/http/`. Add the new mechanism beside it: `delivery/grpc/`, `delivery/queue/`.

Everything bound to a transport (DTOs, protos, endpoints, listeners, event schemas) lives in its own mechanism folder and is wired in `<domain>/<domain>.go`.

**Messaging is always a delivery mechanism.** A queue is a transport in both directions — a consumer is inbound (listens, like HTTP receives requests), a producer is outbound (delivers events to another system). Both live in `delivery/queue/` and both count toward the `delivery/` split: adding any queue to a domain that has HTTP forces the nested layout. Messaging never becomes a `client` integration.

### Repository and client shape

`repository/` — template for any port with multiple backends: interface at the root (`repository/repository.go`), each implementation in its own subpackage (`repository/postgres/`). Same shape for `client/` (port in `client/notifier.go`, adapter in `client/console/`) and for messaging adapters inside `delivery/queue/` (`queue/kafka`, `queue/sqs`).

### External dependencies

Storage, email, cache, third-party APIs — same shape as `client/`, raw connection in `pkg/`, port scoped to who consumes it (domain-specific vs. shared). Messaging is excluded — it's a delivery mechanism, not an integration. See [external-integrations.md](./external-integrations.md).

Wire the module in `<domain>/<domain>.go`. Expose read adapters to other domains in `<domain>/public.go`. Cross-cutting concerns go in `internal/shared` (`core`, `customerr`, `httpx`). Reusable infra goes in `pkg`.

## Related

- [Glossary](./glossary.md) — domain terms with source file links
- [Coding patterns](./coding-patterns.md) — complete implementation reference per layer
- [External integrations](./external-integrations.md) — storage, email, cache, third-party APIs
- [Versioning](./versioning.md) — endpoint-level versioning
