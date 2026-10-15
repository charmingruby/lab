# Cross-module reads

A module never imports another module's internals. To read another module's data, the consumer depends on a port it owns itself. The producer owns two pieces per exposed read, all outside `client/`:

- **`client/` is outbound only** — ports this domain consumes (what it needs from the world). Exposed reads never live here; shape: [external-integrations.md](./external-integrations.md).
- **`<domain>/public/`** — the concrete reader only: a thin struct over a usecase that forwards calls into the consumer's shape. No interface lives here.
- **`<domain>/<domain>.go`** — the assembly: a module-level constructor that builds repositories + use case and returns the concrete `public` reader.

The consumer declares its own port interface (the methods it needs) and codes only against it. The producer's concrete reader satisfies it implicitly. In tests the port is backed by the real public adapter over a real database — never a mock.

## Example — `ticket` exposes `TicketReader` (same shape for any `<domain>`)

**1. The concrete reader** — [internal/ticket/public/ticket_reader.go](../../internal/ticket/public/ticket_reader.go) — no interface, thin struct over a use case:

```go
type TicketReader struct {
	uc *usecase.Usecase
}

func NewTicketReader(uc *usecase.Usecase) *TicketReader {
	return &TicketReader{uc: uc}
}

func (r *TicketReader) GetTicketStatus(ctx context.Context, ticketID string) (string, error) {
	t, err := r.uc.GetTicket(ctx, usecase.GetTicketInput{TicketID: ticketID})
	if err != nil {
		return "", err
	}

	return string(t.Status), nil
}
```

**2. The assembly** — [internal/ticket/ticket.go](../../internal/ticket/ticket.go) — builds repos + use case, returns the concrete reader:

```go
func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, repository.NewTransactionManager(db), notifier.NewConsole())

	return public.NewTicketReader(uc), nil
}
```

**3. The consumer** — declares its own interface, backed in tests by the real adapter over a fresh golden source:

```go
type TicketStatusReader interface {
	GetTicketStatus(ctx context.Context, ticketID string) (string, error)
}

var reader TicketStatusReader
reader, err := ticket.NewTicketReader(db)
// reader.GetTicketStatus(ctx, ticketID) hits the real stack
```

❌ Never import another module's `usecase`, `repository`, or `model` — depend on your own port interface, satisfied by its `public` reader instead.
