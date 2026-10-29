# External integrations (storage, email, cache, third-party APIs)

Any dependency the app doesn't own follows the same three-layer split, regardless of scope. What changes based on scope is only **where the port and adapter live**, never whether they exist.

- **Raw connection**: always in `internal/platform/<capability>/<provider>/`. A thin wrapper around the SDK/client library — connects, configures, exposes primitive operations. No domain awareness, no business logic.
- **Port (interface)**: always defined next to whoever consumes it — either a domain's `client/` or `internal/shared/client/`. This is what usecases depend on. It is never mocked: tests use a real adapter (the production one or an in-memory fake).
- **Adapter**: implements the port by calling the `internal/platform/` raw client. Lives beside the port it implements.

`repository/` never hosts this. External integrations aren't a domain's source of truth — they're consumed, so they're always a `client` port.

A **system-of-record database is not an external integration**. Its toolkit lives in `internal/platform/store/<provider>/` and the domain repository stays concrete — no port. See [platform.md](./platform.md#store).

**Messaging is not an external integration.** A queue is a _delivery mechanism_: its ports and provider adapters live in `internal/platform/messaging/`, and the domain's inbound listener stays in its delivery folder. See [architecture.md](./architecture.md).

This file owns the port-adapter shape. Other docs link here instead of repeating it.

## Shape (canonical)

Contract in `client/<capability>.go` (no `Client` suffix: `Notifier`, not `NotifierClient`), one file per provider in `client/<capability>/`:

```
internal/ticket/client/
├── notifier.go        # port
└── notifier/
    ├── console.go     # production adapter (logs)
    └── memory.go      # test fake (real state + injectable error, never expectations)
```

Source: [internal/ticket/client/](../../internal/ticket/client/)

```go
// internal/ticket/client/notifier.go
type Notifier interface {
	Send(ctx context.Context, input SendNotificationInput) error
}
```

Adapters follow `New<Provider>` (`NewConsole`, `NewMemory`) and are wired in `<domain>/<domain>.go`. The usecase depends on the port, never on the SDK or `internal/platform/` directly. `memory.go` additionally exposes `SetError` / `Sent` for the failure-path test.

## Decision: domain-specific vs. shared

**One question decides where the port and adapter live: does more than one domain need this integration?**

- **One domain** → port + adapters inside that domain's `client/`, as above.
- **Two or more domains** → port + adapters move to `internal/shared/client/<capability>/` (`storage.go` + one file per provider, e.g. `s3.go`). Every consuming domain imports the port from there and wires the concrete adapter in its own `<domain>/<domain>.go`.

The `storage` / `s3` names below are illustrative sketches of the second case:

```go
// sketch, not inventory
type ReceiptStorage interface {
	Upload(ctx context.Context, key string, data []byte) (url string, err error)
}
```

## `internal/platform/` — the raw client, once

The actual SDK call goes through one `internal/platform/<capability>/<provider>/` wrapper, built once and passed down. It never imports from `internal/shared/` or `internal/<domain>/`. Capability taxonomy and layout: [platform.md](./platform.md).

## Testing an integration

The port is never mocked. Each test injects a real implementation, whichever is most viable:

- **The real integration** — a local/sandboxed provider, same philosophy as the golden source via `test/container/`.
- **The `memory.go` fake** — the trivial fallback when a real integration isn't worth it: an in-memory fake with real state and injectable errors, never call expectations.

## Rules

- ✅ Every external integration is a `client` port — never a `repository`. The golden source has no port at all.
- ✅ Raw clients live in `internal/platform/<capability>/<provider>/`, with zero domain awareness.
- ✅ One domain only → port + adapters inside that domain's `client/`.
- ✅ Two or more domains → port + adapter in `internal/shared/client/<capability>/`.
- ✅ Usecases depend on the port interface, never on the `internal/platform/` client or the SDK directly. Tests inject a real implementation (real integration or `memory.go` fake) — never a mock.
- ❌ Never call a `internal/platform/` client directly from a usecase — always through a `client` adapter.
- ❌ Never duplicate the same integration's port across domains once a second domain needs it — move it to `internal/shared/client/`.
- ❌ Never model an external integration as a `repository` — it isn't the domain's source of truth.
- ❌ Never model a queue as a `client` integration — messaging ports and adapters live in `internal/platform/messaging/`.
