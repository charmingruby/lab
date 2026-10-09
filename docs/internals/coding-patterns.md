# Coding Patterns

Implementation reference for a domain feature. `internal/ticket/` is one illustration of the pattern — mirror its shape for any `<domain>`, do not invent new shapes.

> **Module structure**: see [architecture.md](./architecture.md).

## 1. Model

Constructor, invariants, state changes as explicit methods using `core.Model.Touch`.

**Rules:**

- ✅ Constructor returns `(*Model, error)` when there are invariants to hold; plain `*Model` otherwise.
- ✅ Invalid state returns a package-level `var Err...`, not a string.
- ✅ State changes are methods named as verbs (`Assign`, `Resolve`), using `Touch(func(m *core.Model))`.
- ✅ Legitimate values are typed constants with `Valid()`.
- ❌ Never expose raw field mutation from outside — never `ticket.Status = "open"` outside the model package.

Source: [internal/ticket/model/ticket.go](../../internal/ticket/model/ticket.go)

```go
// internal/ticket/model/ticket.go
type TicketStatus string

const (
	OpenTicketStatus       TicketStatus = "open"
	InProgressTicketStatus TicketStatus = "in_progress"
	ResolvedTicketStatus   TicketStatus = "resolved"
)

func NewTicket(input TicketInput) (*Ticket, error) {
	priority := TicketPriority(input.Priority)
	if !priority.Valid() {
		return nil, ErrInvalidPriority
	}

	return &Ticket{
		Model:       core.NewModel(),
		Title:       input.Title,
		Description: input.Description,
		Status:      OpenTicketStatus,
		Priority:    priority,
	}, nil
}

func (t *Ticket) Assign(assigneeID string) error {
	return t.transitionTo(InProgressTicketStatus, func() {
		t.AssigneeID = &assigneeID
	})
}

func (t *Ticket) Resolve() error {
	return t.transitionTo(ResolvedTicketStatus, nil)
}

func (t *Ticket) transitionTo(status TicketStatus, after func()) error {
	if !t.Status.CanTransitionTo(status) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTicketTransition, t.Status, status)
	}

	t.Touch(func(m *core.Model) {
		t.Status = status
		if after != nil {
			after()
		}
	})

	return nil
}
```

---

## 2. Repository (golden source, concrete)

The golden source in `internal/ticket/repository/`. No interface, no `postgres/` subpackage — postgres is effectively immutable, and adding an interface back on the day a swap is needed is trivial.

**Rules:**

- ✅ One concrete struct per aggregate, methods named by business meaning.
- ✅ All methods take `context.Context` first and return the model or count.
- ✅ Not-found returns `(nil, nil)` — never `sql.ErrNoRows`.
- ✅ Query map as `var ticketQueries = map[string]string{ ... }`, prepared once in the constructor.
- ✅ `deleted_at IS NULL` on every read.
- ✅ `context.WithTimeout(ctx, postgrex.DefaultReadTimeout)` per method.
- ✅ `LIMIT $2 OFFSET $3` pagination plus a separate `COUNT(*)` query.
- ❌ No interface for the golden source — usecases take the concrete `*repository.TicketRepository`.
- ❌ No inline queries in methods — always via the prepared `statement(name)`.

Source: [internal/ticket/repository/ticket_repository.go](../../internal/ticket/repository/ticket_repository.go)

```go
// internal/ticket/repository/ticket_repository.go
type TicketRepository struct {
	db    postgrex.Querier
	stmts map[string]*sqlx.Stmt
}

func NewTicketRepository(db postgrex.Querier) (*TicketRepository, error) {
	// prepares every query in ticketQueries once
}

func (r *TicketRepository) FindByID(ctx context.Context, id string) (*model.Ticket, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findTicketByIDQuery)
	if err != nil {
		return nil, err
	}

	var ticket model.Ticket
	if err := stmt.QueryRowxContext(ctx, id).StructScan(&ticket); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &ticket, nil
}
```

Transactions group the concrete repositories in a `Transaction` struct, run through `TransactionManager.Transact` (`postgrex.RunInTx`):

```go
// internal/ticket/repository/transaction_manager.go
type Transaction struct {
	TicketRepo *TicketRepository
}
```

---

## 3. usecase

One file per implemented use case, plus the concrete `Usecase` struct in `usecase/usecase.go`. It takes the concrete repository and transaction manager plus client ports — no mocks anywhere; tests build the real stack.

**Rules:**

- ✅ Input/output structs live in the usecase file, named `<Verb><Resource>Input` / `<Verb><Resource>Output`.
- ✅ Infra failures wrap as `customerr.Integration(err)`.
- ✅ Domain outcomes map: missing → `customerr.NotFound`, duplicate → `customerr.Conflict`, invalid → `customerr.Validation`.
- ✅ Multi-repo writes go through `repository.TransactionManager.Transact`.
- ✅ One `*Usecase` built with `New(...)`, shared by endpoints and public readers.
- ❌ No interface on the usecase — a single implementation needs no port.
- ❌ No HTTP/`net/http` types in the usecase. ❌ No `*sqlx.DB` — the concrete repository only.

Source: [internal/ticket/usecase/](../../internal/ticket/usecase/), [internal/ticket/usecase/usecase.go](../../internal/ticket/usecase/usecase.go)

Simple case — `create_ticket.go`:

```go
func (u *Usecase) CreateTicket(
	ctx context.Context,
	input CreateTicketInput,
) (CreateTicketOutput, error) {
	ticket, err := model.NewTicket(input)
	if err != nil {
		return CreateTicketOutput{}, customerr.Validation(err.Error())
	}

	if err := u.ticketRepo.Create(ctx, ticket); err != nil {
		return CreateTicketOutput{}, customerr.Integration(err)
	}

	return CreateTicketOutput{ID: ticket.ID}, nil
}
```

Transactional case — `assign_ticket.go`:

```go
func (u *Usecase) AssignTicket(
	ctx context.Context,
	input AssignTicketInput,
) error {
	err := u.txManager.Transact(func(tx repository.Transaction) error {
		ticket, err := tx.TicketRepo.FindByID(ctx, input.TicketID)
		if err != nil {
			return customerr.Integration(err)
		}

		if ticket == nil {
			return customerr.NotFound("ticket not found")
		}

		if err := ticket.Assign(input.AssigneeID); err != nil {
			return customerr.Validation(err.Error())
		}

		if err := tx.TicketRepo.Update(ctx, ticket); err != nil {
			return customerr.Integration(err)
		}

		return nil
	})

	return err
}
```

---

## 4. http/endpoint

Parse via `httpx.ParseRequest`, call the use case, answer with `httpx.Write*Response` or `httpx.WriteError`.

**Rules:**

- ✅ A request DTO per endpoint with `validate:` tags (`required,min=1`).
- ✅ Setters answer `httpx.WriteCreatedResponse`, reads `httpx.WriteOKResponse`.
- ✅ Every error goes to `httpx.WriteError` (maps `customerr` type → HTTP status).
- ✅ Path params via `httpx.GetPathParam`.
- ❌ No business logic or repository access in the endpoint. ❌ Don't hand-roll JSON decode/validate.

Source: [internal/ticket/http/endpoint/create_ticket_v1.go](../../internal/ticket/http/endpoint/create_ticket_v1.go)

```go
// internal/ticket/http/endpoint/create_ticket_v1.go
type CreateTicketV1Request struct {
	Title       string `json:"title"       validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	Priority    string `json:"priority"    validate:"required,min=1"`
}

func (e *Endpoint) CreateTicketV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreateTicketV1Request](w, r)
	if err != nil {
		return
	}

	output, err := e.uc.CreateTicket(ctx, usecase.CreateTicketInput{
		Title:       request.Title,
		Description: request.Description,
		Priority:    request.Priority,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreateTicketV1Response{ID: output.ID})
}
```

---

## 5. http/route.go

Register under `/api/v1/...` (the router is mounted at `/api`; group resource routes under `/v1/...`).

**Rules:**

- ✅ `r.Route("/v1/<resources>", ...)` grouping routes per aggregate.
- ✅ `RegisterRoutes` is passed the `*endpoint.Endpoint` built by `SetupEndpoints`.
- ❌ No usecase calls in the router — routing only.

Source: [internal/ticket/http/route.go](../../internal/ticket/http/route.go)

```go
// internal/ticket/http/route.go
func RegisterRoutes(r chi.Router, ep *endpoint.Endpoint) {
	r.Route("/v1/tickets", func(r chi.Router) {
		r.Post("/", ep.CreateTicketV1)
		r.Get("/", ep.ListTicketsV1)
		r.Get("/{id}", ep.GetTicketV1)
		r.Patch("/{id}/assign", ep.AssignTicketV1)
	})
}
```

---

## 6. Wiring — `<domain>/<domain>.go`

Composition root. Build the transaction manager, repositories, usecases, endpoints; register routes.

**Rules:**

- ✅ One function `New(r chi.Router, db *sqlx.DB) error` per module.
- ✅ Concrete repository + client adapters wired here; no mocks anywhere.
- ❌ No wiring in endpoints/usecases — declarative, top-down.

Source: [internal/ticket/ticket.go](../../internal/ticket/ticket.go)

```go
// internal/ticket/ticket.go
func New(r chi.Router, db *sqlx.DB) error {
	txManager := repository.NewTransactionManager(db)

	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return err
	}

	ep := http.SetupEndpoints(
		usecase.New(ticketRepo, txManager, console.NewNotifier()),
	)

	http.RegisterRoutes(r, ep)

	return nil
}
```

Read adapters for other domains go in `<domain>/public.go`:

Source: [internal/ticket/public.go](../../internal/ticket/public.go)

```go
// internal/ticket/public.go
func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, repository.NewTransactionManager(db), console.NewNotifier())

	return public.NewTicketReader(uc), nil
}
```

---

## Structural changes

Not covered here — see [architecture.md](./architecture.md), [versioning.md](versioning.md), and [cross-module-reads.md](cross-module-reads.md). This file only covers the fixed layer order for implementing a single domain feature.
