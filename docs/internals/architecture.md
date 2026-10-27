# Architecture

Ports and adapters. Every domain is a self-contained module with the same dependency spine.

## Request lifecycle

Inbound transport resolves to an endpoint, which follows the domain spine (see [Domain structure](#domain-structure)):

```
delivery → endpoint → usecase → model
                           → repository (concrete golden source)
                           → client port → adapter
```

## Domain wiring

Each domain exposes one composition root in `<domain>.go` — it builds dependencies and hands the router to the endpoint, which registers its own routes. No wiring happens in usecases. Cross-domain read assembly lives in the same file.

Source: [internal/ticket/ticket.go](../../internal/ticket/ticket.go)

## Where code lives

- `cmd/api/main.go` — boots server, mounts domains, starts listening.
- `internal/<domain>/` — one ports-and-adapters module per bounded context.
  - `<domain>.go` — composition root (wires everything, including cross-domain read assembly)
  - `http/endpoint/` — request DTO + handler + versioned route registration (in `New`)
  - `usecase/` — business actions, one file per use case
  - `model/` — entities, invariants, state changes
  - `repository/` — the golden source, concrete (no interface — swapping it is trivial when needed)
  - `client/` — outbound ports this domain consumes. Shape owned by [external-integrations.md](./external-integrations.md). Never hosts exposed reads — those live in `public/`.
  - `public/` — concrete readers for other domains. Shape owned by [cross-module-reads.md](./cross-module-reads.md).
- `internal/shared/` — shared domain language: domain types, typed errors, and ports + adapters used by two or more domains. See the directory for the current set.
- `internal/platform/` — app infrastructure with zero domain awareness: owns cross-cutting capabilities the app needs to run, e.g. config, observability (`logging`, metrics, tracing), transport (`httpx`), provider connections (`postgrex`, `awsx`, `kafka`). See the directory for the current set.
- `pkg/` — reserved for code exposed to the outside world.

## Cross-domain communication

Domains never import each other's `usecase`, `repository`, or `model` packages. The producer exposes a concrete reader in `public/`; the consumer declares its own port interface, satisfied implicitly.

Full rule + example: [cross-module-reads.md](./cross-module-reads.md).

## Error flow

Usecases map domain outcomes to typed errors; endpoints map typed errors to transport statuses. Usecases never return raw strings or unwrapped infrastructure errors.

Source: [internal/shared/customerr/customerr.go](../../internal/shared/customerr/customerr.go), [internal/platform/httpx/response.go](../../internal/platform/httpx/response.go)

## Transaction safety

Writes that must be atomic run through a transaction holding the repositories needed inside it. That covers multi-repo writes and single-aggregate read-modify-write (e.g. load → transition → save). Blind single-repo inserts with no prior read need no explicit transaction.

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

- **One mechanism**: flat, no `delivery/` wrapper. HTTP only → `internal/<domain>/http/`; listener only → `internal/<domain>/listener/`.
- **Two or more mechanisms**: nested — `delivery/` becomes the parent and every mechanism moves inside it: `delivery/http/` beside `delivery/queue/` (listener packages inside), or `delivery/grpc/`. Create it only when a second mechanism appears.

Everything bound to a transport (DTOs, protos, endpoints, listeners, event schemas) lives in its own mechanism folder and is wired in `<domain>/<domain>.go`.

**Messaging is always a delivery mechanism.** A queue is a transport in both directions — a consumer is inbound (listens, like HTTP receives requests), a producer is outbound (delivers events to another system). Both count toward the `delivery/` split: a domain that speaks HTTP and a queue keeps both under `delivery/` (`delivery/http/` + `delivery/queue/`). Messaging never becomes a `client` integration.

Inbound consumers are called **listeners** and are registered in `main`, one goroutine per topic. Shape and rules: [messaging.md](./messaging.md).

### Repository and client shape

`repository/` is concrete — no port interface, no subpackages. For the `client/` port-adapter shape, see [external-integrations.md](./external-integrations.md) (contract in `client/<capability>.go`, one file per provider in `client/<capability>/`). Same capability-per-subpackage shape for messaging adapters inside `delivery/queue/`.

### External dependencies

Storage, email, cache, third-party APIs — same shape as `client/`, raw connection in `internal/platform/`, port scoped to who consumes it (domain-specific vs. shared). Messaging is excluded — it's a delivery mechanism, not an integration. See [external-integrations.md](./external-integrations.md).

Wire the module in `<domain>/<domain>.go`. Expose concrete read adapters to other domains from the same file. Shared domain language goes in `internal/shared/`. Raw clients plus app config go in `internal/platform/`.

## Related

- [Glossary](./glossary.md) — domain terms with source file links
- [Coding patterns](./coding-patterns.md) — complete implementation reference per layer
- [External integrations](./external-integrations.md) — storage, email, cache, third-party APIs
- [Messaging](./messaging.md) — event contracts, publishers, listeners, wiring in `main`
- [Versioning](./versioning.md) — endpoint-level versioning
