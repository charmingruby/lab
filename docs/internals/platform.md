# Platform

App infrastructure with zero domain awareness. `internal/platform/` holds the capabilities the app needs to run: it knows providers and transports, never a domain.

## What belongs

Platform is where external SDKs get wrapped and where process-level concerns live. It is not a layer of the domain spine — a domain reaches it through a `client` port (external integrations), a concrete repository (system of record), or a `messaging` port (publisher).

| Capability  | Shape                       | Layout                                |
| ----------- | --------------------------- | ------------------------------------- |
| `config`    | toolkit                     | `internal/platform/config/`           |
| `httpx`     | toolkit                     | `internal/platform/httpx/`            |
| `o11y`      | toolkit                     | `internal/platform/o11y/<signal>/`    |
| `store`     | toolkit                     | `internal/platform/store/<provider>/` |
| `messaging` | port + adapter + raw client | `internal/platform/messaging/`        |

## Two shapes

Platform capabilities take one of two shapes. Only the first is a port:

1. **Port + adapter + raw client** — the app talks to an external system through a swappable contract. Inside platform this is only `messaging`.
2. **Toolkit** — a thin, provider-specific helper used directly, with no port: `store`, `o11y`, `httpx`, `config`.

Do not invent a port for a toolkit. External integrations are the exception that proves the rule: their port lives in the consuming domain's `client/`, and platform only provides the raw client.

## Store

`store` is the toolkit for the **system of record**, grouped by provider:

```
internal/platform/store/
└── <provider>/
    ├── postgrex.go      # connection (provider entry point)
    ├── statements.go    # prepared statements
    ├── query.go         # query helpers
    ├── transaction.go   # transaction primitive
    └── error.go         # error classification
```

- A new database is a sibling under `internal/platform/store/<provider>/`, never a new top-level capability.
- There is no `store` port. The golden source is concrete, so the repository imports exactly one store toolkit.
- Swapping databases is a repository rewrite, not an adapter swap.
- `Statements`, `One`, `Many`, `Exec`, and `ExecReturning` are SQL-specific. A non-SQL store exposes its own primitives.
- The transaction boundary stays in the domain's `repository/` (`TransactionManager`). `store` only provides the primitive (`RunInTx`).
- A cache, email provider, object storage, or third-party API is **not** a store: it is a `client` integration with a raw client under `internal/platform/<capability>/<provider>/`.

Repository usage: [coding-patterns.md](./coding-patterns.md#2-repository-golden-source-concrete).

## Observability

All signals live under `o11y/`, one package per signal:

```
internal/platform/o11y/
└── <signal>/
```

`logging` is the signal that exists today; `metrics` and `tracing` join it as sibling packages when the app emits those signals.

- `o11y` is a toolkit, not a port. Consumers call it directly (logger from context, tracer, meter).
- A signal with a specific backend puts it under `o11y/<signal>/<provider>/`, keeping the signal package as the entry point.
- Observability never imports a domain and never holds business rules.

## HTTP

`httpx/` is the transport toolkit:

```
internal/platform/httpx/
├── server.go       # server, middleware, probes
├── request.go      # parse + validate
├── response.go     # success and error writers
├── validator.go
└── client.go       # outbound HTTP client
```

It carries no domain knowledge: handlers, DTOs, and route registration belong to the domain.

## Config

- `internal/platform/config/` loads typed configuration from the environment once, at boot.
- Defaults and validation live with the config; invalid or missing required configuration aborts startup.
- One config per app. Shared platform packages never import it.
- It exposes primitives (URL, port, secret, TTL), never a domain object.

## Messaging

```
internal/platform/messaging/
├── publisher.go              # port
├── listener.go               # port
├── publisher/
│   └── <provider>.go         # outbound adapter
├── listener/
│   └── <provider>.go         # inbound adapter: decode + call Handle
└── broker/
    └── <provider>/           # raw SDK client
        ├── reader.go
        ├── writer.go
        └── message.go
```

- Ports stay at the root. No provider type crosses the port boundary.
- Adapters translate between the port and the raw client. One file per provider per direction.
- `broker/` is not a port and not an adapter: it is the only place that imports the broker SDK and owns connection, fetch/commit, retry, and backoff.
- A new broker adds `broker/<provider>/`, `publisher/<provider>.go`, and `listener/<provider>.go`. The ports do not change.
- The domain's listener stays in the domain. The adapter decodes the message and calls `Handle`.

Semantics of publishers, listeners, and event contracts: [messaging.md](./messaging.md).

## Adding a capability

| You need...                            | It goes in                                                                              |
| -------------------------------------- | --------------------------------------------------------------------------------------- |
| a new system-of-record database        | `internal/platform/store/<provider>/` plus a concrete repository                        |
| cache, email, storage, third-party API | `internal/platform/<capability>/<provider>/` raw client plus a domain `client/` port    |
| a new broker                           | `internal/platform/messaging/broker/<provider>/` plus the two adapters; ports unchanged |
| a new observability signal             | `internal/platform/o11y/<signal>/`                                                      |
| new HTTP behavior                      | `internal/platform/httpx/`                                                              |
| new configuration or env var           | `internal/platform/config/`                                                             |

## Rules

- ✅ Platform has zero domain awareness.
- ✅ Use a port only when there is a swappable external contract.
- ✅ Keep provider-specific code in the provider leaf; new providers do not change ports.
- ✅ Keep the system of record in `store/`; keep consumed integrations in `client/`.
- ❌ No model, usecase, repository, listener, or business rule in platform.
- ❌ No import of a domain package.
- ❌ Never name a package, type, or file `subscriber` or `consumer`.
- ❌ Never put a cache or third-party integration under `store/`.

## Related

- [Architecture](./architecture.md) — where platform sits in the dependency spine
- [Messaging](./messaging.md) — publishers, listeners, event contracts
- [External integrations](./external-integrations.md) — ports for consumed integrations
- [Coding patterns](./coding-patterns.md) — repository usage over `store`
