# Glossary

Living index of terms. One line per term + where the rule lives. Rules live in their owner doc — this file never duplicates them.

## Table of contents

- [Domain structure](#domain-structure)
- [Ports and adapters](#ports-and-adapters)
- [Delivery mechanisms](#delivery-mechanisms)
- [Cross-cutting](#cross-cutting)

## Domain structure

#### Domain

A bounded context in `internal/<domain>/`. Never imports another domain's internals. Rule: [architecture.md](./architecture.md), example: [internal/ticket/](../../internal/ticket/)

#### Model

Entity with constructor invariants + verb state-change methods. Rule: [coding-patterns.md](./coding-patterns.md#1-model)

#### Usecase

One business action; orchestrates model + repository + client ports. Rule: [coding-patterns.md](./coding-patterns.md#3-usecase)

#### Repository

Concrete golden source over prepared statements and the `postgrex` helpers; no interface. Rule: [coding-patterns.md](./coding-patterns.md#2-repository-golden-source-concrete)

#### Client (outbound port)

Interface a domain consumes for external systems. Shape: [external-integrations.md](./external-integrations.md#shape-canonical)

#### Adapter

Concrete implementation behind a client port. Shape: [external-integrations.md](./external-integrations.md#shape-canonical)

#### Endpoint

Handler that parses, calls usecase, writes response. Rule: [coding-patterns.md](./coding-patterns.md#4-httpendpoint)

## Ports and adapters

#### Port

Interface decoupling usecase from externals; outbound only. Shape: [external-integrations.md](./external-integrations.md#shape-canonical)

#### Delivery mechanism

Any transport a domain speaks (HTTP, gRPC, queue). Rule: [architecture.md](./architecture.md#delivery-mechanism-layout)

#### Dependency spine

Layered flow within a domain. Canonical diagram: [architecture.md](./architecture.md#domain-structure)

#### Public adapter

Concrete reader over a usecase for cross-domain reads; consumer owns the interface. Rule: [cross-module-reads.md](./cross-module-reads.md)

## Delivery mechanisms

#### Flat layout

Default: one inbound mechanism, no wrapper. `http/endpoint/` with route registration in `New`, or `listener/on_<event>_v<n>/` for queue-only. An outbound publisher keeps the module flat. Rule: [architecture.md](./architecture.md#delivery-mechanism-layout)

#### Nested layout

Two or more inbound mechanisms under `delivery/` — e.g. `delivery/http/` + `delivery/queue/`. Only exists when a second inbound transport appears. Rule: [architecture.md](./architecture.md#delivery-mechanism-layout)

#### Listener

Inbound queue consumer: `listener/on_<event>_v<n>/` with `Listener`, `New`, `Handle`; registered in `main`, one goroutine per topic. Rule: [messaging.md](./messaging.md#listener-inbound)

#### Publisher

Outbound queue writer, injected as the `Publisher` port. Rule: [messaging.md](./messaging.md#publisher-outbound)

#### Event contract

Versioned package holding the topic name + message struct, shared by publisher and listener. Rule: [messaging.md](./messaging.md#event-contracts)

#### Transaction manager

Atomic unit for multi-repo writes and read-modify-write. Rule: [architecture.md](./architecture.md#transaction-safety)

## Cross-cutting

#### shared

Domain types, typed errors, ports + adapters used by 2+ domains. See directory: [internal/shared/](../../internal/shared/)

#### platform

App infrastructure with zero domain awareness: `config`, `httpx`, `o11y`, `store`, and `messaging`. Only messaging is port + adapter. Rule: [platform.md](./platform.md)

#### store

System-of-record toolkit under `internal/platform/store/<provider>/`; no port, and the repository stays concrete. Rule: [platform.md](./platform.md#store)

#### config

Typed environment configuration loaded at boot from `internal/platform/config/`. Rule: [platform.md](./platform.md#config)

#### pkg (reserved)

Code exposed to the outside world. Absent here — not for internal infra.

## Practical shortcuts

- If you see `port`, think "interface against an external provider."
- If you see `adapter`, think "concrete implementation behind a client port."
- If you see `usecase`, think "one business action, orchestrates everything."
- If you see `delivery`, think "wrapper that only exists when the domain speaks two or more transports (e.g. `delivery/http/` + `delivery/queue/`)."
- If you see `listener`, think "inbound queue consumer registered in `main` — flat at `<domain>/listener/`, or `<domain>/delivery/queue/` once HTTP is also present."
- If you see `<domain>.go`, think "module wiring + cross-domain read assembly."
