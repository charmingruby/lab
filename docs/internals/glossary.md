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

The domain entity with constructor invariants and state-change methods. The source of truth for business rules. Models use `core.Model` for base fields (ID, timestamps) and `core.Model.Touch` for state mutations. Constructors return `(*Model, error)` when there are invariants to enforce.

Source: [internal/ticket/model/ticket.go](../../internal/ticket/model/ticket.go), [internal/shared/core/model.go](../../internal/shared/core/model.go)

#### Usecase

The application service that orchestrates one business action. It coordinates models, repositories, and clients. Input/output structs live in the usecase file (`<Verb><Resource>Input` / `<Verb><Resource>Output`). Usecases never import HTTP or SQL types.

Source: [internal/ticket/usecase/](../../internal/ticket/usecase/), [internal/ticket/usecase/usecase.go](../../internal/ticket/usecase/usecase.go)

#### Repository

The persistence port (Go `interface`) for a domain's own data. One interface per aggregate, methods named by business meaning. All methods take `context.Context` first. Not-found returns `(nil, nil)`, never `sql.ErrNoRows`.

Source: [internal/ticket/repository/repository.go](../../internal/ticket/repository/repository.go)

#### Client (outbound port)

An outbound port — the interface a domain exposes for others to read its data, or the interface a domain consumes to call external systems. All ports for a domain (outbound and exposed reads) live together in `client/*.go`.

Source: [internal/ticket/client/](../../internal/ticket/client/)

#### Adapter

The concrete implementation behind a port. Postgres is an adapter for the repository port; a console notifier is an adapter for the notification client port. Adapters live beside their port in a subdirectory.

Source: [internal/ticket/repository/postgres/](../../internal/ticket/repository/postgres/), [internal/ticket/client/console/](../../internal/ticket/client/console/)

#### Endpoint

The HTTP handler that parses a request (via `httpx.ParseRequest`), calls a usecase, and writes a response (via `httpx.Write*Response`). One DTO per endpoint with `validate:` tags. Endpoints never contain business logic.

Source: [internal/ticket/http/endpoint/](../../internal/ticket/http/endpoint/)

## Ports and adapters

#### Port

A Go `interface` that decouples a layer from its implementation. Used for both inbound boundaries (repository) and outbound boundaries (client). Ports are defined at the root of their directory (e.g., `repository/repository.go`).

Source: [internal/ticket/repository/repository.go](../../internal/ticket/repository/repository.go), [internal/ticket/client/notifier.go](../../internal/ticket/client/notifier.go)

#### Delivery mechanism

Any transport a domain speaks: HTTP, gRPC, queue consumer, queue producer. One mechanism = flat layout (`http/endpoint/`, `http/route.go`). Two or more = nested under `delivery/`. Messaging (queues) is always a delivery mechanism, never a client integration.

Source: [architecture.md](./architecture.md)

#### Dependency spine

The layered dependency flow within a domain:

```
<protocol> → usecase → repository (port) → repository/postgres
                              → client (port)   → client/console
                              → model
```

Data flows right-to-left: the adapter implements the port, the usecase depends on the port, the protocol (endpoint) depends on the usecase. Nothing crosses layers in the wrong direction.

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

Wraps `postgrex.RunInTx` for multi-repo writes. Typed as `core.TransactionManager[repository.Transaction]`. The `Transaction` struct holds the repo interfaces needed inside the transaction.

Source: [internal/shared/core/transaction.go](../../internal/shared/core/transaction.go), [internal/ticket/repository/postgres/transaction_manager.go](../../internal/ticket/repository/postgres/transaction_manager.go)

## Cross-cutting

#### core

Shared base utilities in `internal/shared/core/`: `Model` (base entity with ID and timestamps), `PaginationParams`, `TransactionManager`. Every domain depends on `core`.

Source: [internal/shared/core/](../../internal/shared/core/)

#### customerr

Typed error package in `internal/shared/customerr/`. Maps domain outcomes to HTTP status codes: `NotFound` → 404, `Conflict` → 409, `Validation` → 422, `Integration` → 500. Usecases wrap infra failures as `customerr.Integration(err)`.

Source: [internal/shared/customerr/customerr.go](../../internal/shared/customerr/customerr.go)

#### httpx

HTTP utilities in `internal/shared/httpx/`: `ParseRequest[T]`, `WriteOKResponse`, `WriteCreatedResponse`, `WriteError`, `GetPathParam`. Endpoints use these instead of hand-rolling JSON decode/validate.

Source: [internal/shared/httpx/](../../internal/shared/httpx/)

## Practical shortcuts

- If you see `port`, think "interface that decouples a layer."
- If you see `adapter`, think "concrete implementation behind a port."
- If you see `usecase`, think "one business action, orchestrates everything."
- If you see `delivery`, think "transport the domain speaks (HTTP, gRPC, queue)."
- If you see `public.go`, think "cross-domain read assembly."
