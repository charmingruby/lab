# External integrations (storage, email, cache, third-party APIs)

Any dependency the app doesn't own — object storage, email delivery, cache, a third-party API — follows the same three-layer split, regardless of scope. What changes based on scope is only **where the port and adapter live**, never whether they exist.

- **Raw connection**: always in `pkg/`. A thin wrapper around the SDK/client library — connects, configures, exposes primitive operations. No domain awareness, no business logic.
- **Port (interface)**: always defined next to whoever consumes it — either a domain's `client/` or `internal/shared/client/`. This is what usecases actually depend on and what gets mocked.
- **Adapter**: implements the port by calling the `pkg/` raw client. Lives beside the port it implements.

`repository/` never hosts this. `repository/` is reserved for the domain's own persistent golden source of truth — the store that owns and mutates the domain's state — deliberately kept local to the domain, one interface at the root, one subpackage per backend (`repository/postgres` today, another tomorrow if the backend changes). External integrations aren't a domain's source of truth; they're consumed, so they're always a `client` port, whether domain-specific or shared.

## Decision: domain-specific vs. shared

**One question decides where the port and adapter live: does more than one domain need this integration?**

### Used by a single domain

Port and adapter live inside that domain, same shape as any other `client/` port (e.g. `NotificationClient`):

```
internal/billing/
├── client/
│   ├── notifier.go        # existing outbound port
│   ├── storage.go         # new port: interface only
│   └── s3/
│       └── s3.go          # adapter: implements storage.go, wraps pkg/s3
```

```go
// internal/billing/client/storage.go
type ReceiptStorage interface {
	Upload(ctx context.Context, key string, data []byte) (url string, err error)
}
```

```go
// internal/billing/client/s3/s3.go
type ReceiptStorage struct {
	raw *s3pkg.Client // pkg/s3 raw wrapper
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

Port and adapter move to `internal/shared/client/<capability>/`, one subpackage per provider — same `repository/`-style shape, just for a cross-domain port instead of a cross-backend one:

```
internal/shared/
└── client/
    └── storage/
        ├── storage.go      # port: interface only
        └── s3/
            └── s3.go       # adapter: implements storage.go, wraps pkg/s3
```

```go
// internal/shared/client/storage/storage.go
type Storage interface {
	Upload(ctx context.Context, bucket, key string, data []byte) (url string, err error)
}
```

```go
// internal/shared/client/storage/s3/s3.go
type Storage struct {
	raw *s3pkg.Client
}

func New(raw *s3pkg.Client) *Storage {
	return &Storage{raw: raw}
}

func (s *Storage) Upload(ctx context.Context, bucket, key string, data []byte) (string, error) {
	return s.raw.PutObject(ctx, bucket, key, data)
}
```

Every consuming domain imports the port from `internal/shared/client/storage`, and wires the concrete adapter (`storage/s3`) in its own `<domain>/<domain>.go` — the domain still only ever codes against the interface.

```go
// internal/billing/billing.go
storage := s3.New(s3pkg.NewClient(cfg))
createReceipt := usecase.NewCreateReceiptUsecase(storage) // typed as shared client.Storage
```

## `pkg/` — the raw client, once

Regardless of domain-specific or shared scope, the actual SDK call goes through one `pkg/` wrapper, built once and passed down to whichever adapter needs it:

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

`pkg/` never imports anything from `internal/` — it has no notion of domains, ports, or business rules. This is the same relationship `pkg/postgrex` has with `repository/postgres`: one raw wrapper, reused by every adapter that needs a database connection.

## Rules

- ✅ Every external integration is a `client` port — never a `repository`.
- ✅ Raw SDK/connection code lives in `pkg/`, with zero domain awareness.
- ✅ One domain only → port + adapter inside that domain's `client/`.
- ✅ Two or more domains → port + adapter in `internal/shared/client/<capability>/<provider>/`.
- ✅ Usecases depend on the port interface, never on the `pkg/` client or the SDK directly.
- ❌ Never call a `pkg/` client directly from a usecase — always through a `client` adapter.
- ❌ Never duplicate the same integration's port across domains once a second domain needs it — move it to `internal/shared/client/` instead of copy-pasting.
- ❌ Never model an external integration as a `repository` — it isn't the domain's source of truth.

BAD — usecase calling the SDK wrapper directly, skipping the port:

```go
// ❌ Usecase importing pkg/s3 directly
import "github.com/charmingruby/new/pkg/s3"

func (u *createReceiptUsecase) CreateReceipt(ctx context.Context, input CreateReceiptInput) error {
	client := s3.NewClient(u.cfg)          // ❌ no port, no mock surface
	client.PutObject(ctx, "receipts", ...) // ❌ usecase now knows about S3
	...
}

// ✅ Usecase depends on the port, constructed and injected at the module root
func (u *createReceiptUsecase) CreateReceipt(ctx context.Context, input CreateReceiptInput) error {
	url, err := u.storage.Upload(ctx, input.Key, input.Data) // client.ReceiptStorage or shared client.Storage
	...
}
```
