# Glossary

This is a living glossary for the Go backend template. It explains what common terms mean in this codebase and where the code lives.

## Table of contents

- [Domain structure](#domain-structure)
- [Ports and adapters](#ports-and-adapters)
- [Delivery mechanisms](#delivery-mechanisms)
- [Cross-cutting](#cross-cutting)

## Domain structure

#### Domain

A bounded context in `internal/<domain>/`. Each domain is a self-contained module with its own model, usecase, repository, and delivery mechanism(s). Domains never import each other's internals — they communicate through client ports.

Source: [architecture.md](./architecture.md), [internal/ticket/](../../internal/ticket/)

#### Model

The domain entity with constructor invariants and state-change methods. The source of truth for business rules. Models build on the shared base model (currently `core.Model`) for base fields (ID, timestamps) and state-mutation helpers (currently `core.Model.Touch`). Constructors return `(*Model, error)` when there are invariants to enforce.

Source: [internal/ticket/model/ticket.go](../../internal/ticket/model/ticket.go), [internal/shared/core/model.go](../../internal/shared/core/model.go)

#### Usecase

The application service that orchestrates one business action. It coordinates models, repositories, and clients. Input/output structs live in the usecase file (`<Verb><Resource>Input` / `<Verb><Resource>Output`). Usecases never import HTTP or SQL types.

Source: [internal/ticket/usecase/](../../internal/ticket/usecase/), [internal/ticket/usecase/usecase.go](../../internal/ticket/usecase/usecase.go)

#### Repository

The golden source for a domain's own data — a concrete postgres struct, not an interface. One struct per aggregate, methods named by business meaning. All methods take `context.Context` first. Not-found returns `(nil, nil)`, never `sql.ErrNoRows`. There is no port here on purpose: postgres is effectively immutable, and re-introducing an interface on the day a swap is needed is trivial.

Source: [internal/ticket/repository/ticket_repository.go](../../internal/ticket/repository/ticket_repository.go)

#### Client (outbound port)

An outbound port — the interface a domain exposes for others to read its data, or the interface a domain consumes to call external systems. Ports exist here as an anti-corruption layer: external providers are easily swapped. All ports for a domain (outbound and exposed reads) live together in `client/*.go`.

Source: [internal/ticket/client/](../../internal/ticket/client/)

#### Adapter

The concrete implementation behind a client port. A console notifier is an adapter for the notification client port; an in-memory notifier is the test adapter for the same port. Adapters live beside their port in a subdirectory. The repository is not an adapter — it has no port.

Source: [internal/ticket/client/console/](../../internal/ticket/client/console/), [internal/ticket/client/memory/](../../internal/ticket/client/memory/)

#### Endpoint

The HTTP handler that parses a request (via the shared HTTP helpers, currently `httpx.ParseRequest`), calls a usecase, and writes a response (via the shared HTTP writers, currently `httpx.Write*Response`). One DTO per endpoint with `validate:` tags. Endpoints never contain business logic.

Source: [internal/ticket/http/endpoint/](../../internal/ticket/http/endpoint/)

## Ports and adapters

#### Port

A Go `interface` that decouples the usecase from an external implementation. Used only for outbound boundaries (`client/`): external providers are easily swapped, so that anti-corruption layer pays off. The golden source has no port — `repository/` is concrete.

Source: [internal/ticket/client/notifier.go](../../internal/ticket/client/notifier.go)

#### Delivery mechanism

Any transport a domain speaks: HTTP, gRPC, queue consumer, queue producer. One mechanism = flat layout (`http/endpoint/`, `http/route.go`). Two or more = nested under `delivery/`. Messaging (queues) is always a delivery mechanism, never a client integration.

Source: [architecture.md](./architecture.md)

#### Dependency spine

The layered dependency flow within a domain:

```
<protocol> → usecase → repository (concrete postgres)
                      → client (port)   → client/console
                      → model
```

Data flows right-to-left: the usecase calls the concrete repository and the client ports, the protocol (endpoint) depends on the usecase. Nothing crosses layers in the wrong direction.

Source: [architecture.md](./architecture.md), [internal/ticket/ticket.go](../../internal/ticket/ticket.go)

#### Public adapter

A thin struct over a usecase that exposes a client port for cross-domain reads. Assembled in `<domain>/public.go`, which builds repositories + usecase and returns the adapter typed as the client port.

Source: [internal/ticket/public.go](../../internal/ticket/public.go), [internal/ticket/public/ticket_reader.go](../../internal/ticket/public/ticket_reader.go)

## Delivery mechanisms

#### Flat layout

The default when a domain has only one delivery mechanism (typically HTTP). `internal/<domain>/http/` holds `endpoint/` and `route.go`. No `delivery/` wrapper.

Source: [internal/ticket/http/](../../internal/ticket/http/)

#### Nested layout

Used when a domain has two or more delivery mechanisms. Introduce `delivery/` as the parent. Move existing `http/` to `delivery/http/`. Add new mechanisms beside it: `delivery/grpc/`, `delivery/queue/`.

Source: [architecture.md](./architecture.md)

#### Transaction manager

Wraps the shared DB helper (currently `postgrex.RunInTx`) for multi-repo writes. A concrete `repository.TransactionManager` taking a `repository.Transaction` with the repositories needed inside the transaction.

Source: [internal/ticket/repository/transaction_manager.go](../../internal/ticket/repository/transaction_manager.go)

## Cross-cutting

#### shared

Shared domain language in `internal/shared/`: domain types, typed errors, and ports + adapters used by two or more domains. Usual members are a base model with ID/timestamps, pagination params, a transaction manager type, and typed errors mapping domain outcomes to HTTP status (`NotFound` → 404, `Conflict` → 409, `Validation` → 422, `Integration` → 500). See the directory for the current set — do not treat this doc as inventory.

Source: [internal/shared/](../../internal/shared/)

#### platform

Internal infrastructure in `internal/platform/` with zero domain awareness: raw external clients and transport/config helpers (e.g. HTTP, logging, DB, validation). Adapters in `internal/shared/client/` or a domain's `client/` wrap it. `internal/platform/` never imports from `internal/shared/` or `internal/<domain>/`. See the directory for the current set.

Source: [internal/platform/](../../internal/platform/)

#### pkg (reserved)

Reserved for code exposed to the outside world (e.g. public API contract). Absent in this repo — do not use it for internal infra.

## Practical shortcuts

- If you see `port`, think "interface against an external provider."
- If you see `adapter`, think "concrete implementation behind a client port."
- If you see `usecase`, think "one business action, orchestrates everything."
- If you see `delivery`, think "transport the domain speaks (HTTP, gRPC, queue)."
- If you see `public.go`, think "cross-domain read assembly."
