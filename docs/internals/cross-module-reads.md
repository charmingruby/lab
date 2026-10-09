# Cross-module reads

A module never imports another module's use case. To read another module's data, the consumer depends on a **client port** owned by the producer. The producer owns three pieces per exposed read:

- **`<domain>/client/`** — the port: a read interface the consumer codes against (e.g. `TicketReader.GetTicket`). Also holds outbound ports the module itself consumes (e.g. `NotificationClient`, adapter in `client/console`). All ports for a domain — outbound and exposed reads — live together in `client/*.go`; split into separate files only when a single file grows unwieldy.
- **`<domain>/public/`** — the adapter: a thin struct over a use case that forwards calls into the port shape (e.g. `public.NewTicketReader(getTicket)`, whose `GetTicketStatus` delegates to `getTicket.GetTicket`).
- **`<domain>/public.go`** — the assembly: a module-level constructor (e.g. `NewTicketReader(db)`) that builds repositories + use case and returns the adapter typed as the `client` port.

The consumer codes only against the produced `client` port. In tests the port is backed by the real public adapter over a real database — never a mock.

## Example — `ticket` exposes `TicketReader` (illustration — same shape for any `<domain>`)

**1. The port** — [internal/ticket/client/notifier.go](../../internal/ticket/client/notifier.go):

```go
type TicketReader interface {
	GetTicketStatus(ctx context.Context, ticketID string) (string, error)
}
```

**2. The adapter** — [internal/ticket/public/ticket_reader.go](../../internal/ticket/public/ticket_reader.go) — thin struct over a use case:

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

**3. The assembly** — [internal/ticket/public.go](../../internal/ticket/public.go) — builds repos + use case, returns the adapter typed as the port:

```go
func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, repository.NewTransactionManager(db), console.NewNotifier())

	return public.NewTicketReader(uc), nil
}
```

**4. The consumer** — depends only on the client port, backed in tests by the real adapter (`ticket.NewTicketReader(db)` over `container.StartPostgres(t)`):

```go
reader, err := ticket.NewTicketReader(container.StartPostgres(t))
// reader.GetTicketStatus(ctx, ticketID) hits the real stack
```

❌ Never import another module's `usecase`, `repository`, or `model` — depend on its `client` port instead.
