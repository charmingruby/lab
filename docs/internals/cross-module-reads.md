# Cross-module reads

A module never imports another module's internals. To read another module's data, the consumer depends on a port owned by the producer. The producer owns three pieces per exposed read, all outside `client/`:

- **`client/` is outbound only** — ports this domain consumes (what it needs from the world: notifier, storage). Exposed reads never live here.
- **`<domain>/public/`** — the port + the adapter: the read interface plus a thin struct over a usecase that forwards calls into the port shape.
- **`<domain>/public.go`** — the assembly: a module-level constructor that builds repositories + use case and returns the adapter typed as the `public` port.

The consumer codes only against the produced `public` port. In tests the port is backed by the real public adapter over a real database — never a mock.

## Example — `ticket` exposes `Reader` (current illustration — same shape for any `<domain>`)

**1. The port + 2. the adapter** — [internal/ticket/public/ticket_reader.go](../../internal/ticket/public/ticket_reader.go) — port beside its implementation, thin struct over a use case:

```go
type Reader interface {
	GetTicketStatus(ctx context.Context, ticketID string) (string, error)
}

type TicketReader struct {
	uc *usecase.Usecase
}

func NewTicketReader(uc *usecase.Usecase) Reader {
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
func NewTicketReader(db *sqlx.DB) (public.Reader, error) {
	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, repository.NewTransactionManager(db), console.NewNotifier())

	return public.NewTicketReader(uc), nil
}
```

**4. The consumer** — depends only on the public port, backed in tests by the real adapter over a fresh golden source:

```go
import "github.com/charmingruby/lab/internal/ticket/public"

var reader public.Reader
reader, err := ticket.NewTicketReader(db)
// reader.GetTicketStatus(ctx, ticketID) hits the real stack
```

❌ Never import another module's `usecase`, `repository`, or `model` — depend on its `public` port instead.

- Current: `db` is `*sqlx.DB` via `container.StartPostgres(t)`.
