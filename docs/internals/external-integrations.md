# External integrations (storage, email, cache, third-party APIs)

Any dependency the app doesn't own — object storage, email delivery, cache, a third-party API — follows the same three-layer split, regardless of scope. What changes based on scope is only **where the port and adapter live**, never whether they exist.

- **Raw connection**: always in `pkg/`. A thin wrapper around the SDK/client library — connects, configures, exposes primitive operations. No domain awareness, no business logic.
- **Port (interface)**: always defined next to whoever consumes it — either a domain's `client/` or `internal/shared/client/`. This is what usecases depend on and what gets mocked.
- **Adapter**: implements the port by calling the `pkg/` raw client. Lives beside the port it implements.

`repository/` never hosts this. External integrations aren't a domain's source of truth — they're consumed, so they're always a `client` port.

**Messaging is not an external integration.** A queue is a *delivery mechanism* and lives in `delivery/queue/`. See [architecture.md](./architecture.md).

## Decision: domain-specific vs. shared

**One question decides where the port and adapter live: does more than one domain need this integration?**

### Used by a single domain

Port and adapter live inside that domain, same shape as any other `client/`:

```
internal/ticket/
├── client/
│   ├── notifier.go        # outbound port
│   └── storage.go         # new port: interface only
```

Source: [internal/ticket/client/](../../internal/ticket/client/)

```go
// internal/ticket/client/storage.go
type ReceiptStorage interface {
	Upload(ctx context.Context, key string, data []byte) (url string, err error)
}
```

The adapter wraps `pkg/` raw client, lives in `client/<provider>/`:

```go
// internal/ticket/client/s3/s3.go
type ReceiptStorage struct {
	raw *s3pkg.Client
}

func NewReceiptStorage(raw *s3pkg.Client) *ReceiptStorage {
	return &ReceiptStorage{raw: raw}
}

func (s *ReceiptStorage) Upload(ctx context.Context, key string, data []byte) (string, error) {
	return s.raw.PutObject(ctx, key, data)
}
```

The usecase depends on `client.ReceiptStorage`, never on `s3pkg.Client` directly.

### Used by two or more domains

Port and adapter move to `internal/shared/client/<capability>/`, one subpackage per provider:

```
internal/shared/
└── client/
    └── storage/
        ├── storage.go      # port: interface only
        └── s3/
            └── s3.go       # adapter: implements storage.go, wraps pkg/s3
```

Every consuming domain imports the port from `internal/shared/client/storage`, and wires the concrete adapter in its own `<domain>/<domain>.go`:

```go
// internal/ticket/ticket.go
storage := s3.New(s3pkg.NewClient(cfg))
createReceipt := usecase.NewCreateReceiptUsecase(storage) // typed as shared client.Storage
```

## `pkg/` — the raw client, once

The actual SDK call goes through one `pkg/` wrapper, built once and passed down:

```go
// pkg/s3/s3.go
type Client struct {
	sdk *awss3.Client
}

func NewClient(cfg Config) *Client { /* ... */ }

func (c *Client) PutObject(ctx context.Context, bucket, key string, data []byte) (string, error) {
	// raw SDK call, no domain types
}
```

`pkg/` never imports anything from `internal/`.

## Rules

- ✅ Every external integration is a `client` port — never a `repository`.
- ✅ Raw SDK/connection code lives in `pkg/`, with zero domain awareness.
- ✅ One domain only → port + adapter inside that domain's `client/`.
- ✅ Two or more domains → port + adapter in `internal/shared/client/<capability>/<provider>/`.
- ✅ Usecases depend on the port interface, never on the `pkg/` client or the SDK directly.
- ❌ Never call a `pkg/` client directly from a usecase — always through a `client` adapter.
- ❌ Never duplicate the same integration's port across domains once a second domain needs it — move it to `internal/shared/client/`.
- ❌ Never model an external integration as a `repository` — it isn't the domain's source of truth.
- ❌ Never model a queue as a `client` integration — messaging lives in `delivery/queue/`.
