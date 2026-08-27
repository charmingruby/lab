# Coding Patterns

Implementation reference for a domain feature. Real code from `internal/ticket/`. Mirror it — do not invent new shapes.

> **Module structure**: see "Structure" in `AGENTS.md`.

## 1. Model

Constructor, invariants, state changes as explicit methods using `core.Model.Touch`.

**Rules:**

- ✅ Constructor returns `(*Model, error)` when there are invariants to hold; plain `*Model` otherwise.
- ✅ Invalid state returns a package-level `var Err...`, not a string.
- ✅ State changes are methods named as verbs (`Assign`, `Resolve`), using `Touch(func(m *core.Model))`.
- ✅ Legitimate values are typed constants with `Valid()`.
- ❌ Never expose raw field mutation from outside — ❌ never `ticket.Status = "open"` outside the model package.

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

## 2. Repository (port)

Interface at the root: `internal/ticket/repository/repository.go`.

**Rules:**

- ✅ One interface per aggregate, methods named by business meaning.
- ✅ All methods take `context.Context` first and return the model or count.
- ✅ Not-found returns `(nil, nil)` — never `sql.ErrNoRows`.
- ❌ No ORM/SQL types in the port — usecases and mocks stay transport-free.

```go
// internal/ticket/repository/repository.go
type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	FindByID(ctx context.Context, id string) (*model.Ticket, error)
	Update(ctx context.Context, ticket *model.Ticket) error
	ListByStatus(ctx context.Context, status string, params core.PaginationParams) ([]model.Ticket, int, error)
}

type Transaction struct {
	TicketRepo TicketRepository
}
```

---

## 3. repository/postgres

Prepared statements via `postgrex.Querier`.

**Rules:**

- ✅ Query map as `var ticketQueries = map[string]string{ ... }`, prepared once in the constructor.
- ✅ `deleted_at IS NULL` on every read.
- ✅ `context.WithTimeout(ctx, postgrex.DefaultReadTimeout)` per method.
- ✅ Not-found returns `(nil, nil)` — never surface `sql.ErrNoRows` to callers.
- ✅ `LIMIT $2 OFFSET $3` pagination plus a separate `COUNT(*)` query.
- ❌ No inline queries in methods — always via the prepared `statement(name)`.

```go
// internal/ticket/repository/postgres/ticket_repository.go
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

---

## 4. usecase

One file per implemented use case; interfaces aggregated in `usecase/usecase.go` (what mockery mocks).

**Rules:**

- ✅ Input/output structs live in the usecase file, named `<Verb><Resource>Input` / `<Verb><Resource>Output`.
- ✅ Infra failures wrap as `customerr.Integration(err)`.
- ✅ Domain outcomes map: missing → `customerr.NotFound`, duplicate → `customerr.Conflict`, invalid → `customerr.Validation`.
- ✅ Multi-repo writes go through `core.TransactionManager[repository.Transaction]` (`postgrex.RunInTx`).
- ✅ Interface stays in `usecase.go`, typed by the concrete `*usecase` constructing it.
- ❌ No HTTP/`net/http` types in the usecase. ❌ No `*sqlx.DB` — repositories only.

Simple case — `create_ticket.go`:

```go
func (u *createTicketUsecase) CreateTicket(
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
func (u *assignTicketUsecase) AssignTicket(
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

## 5. http/endpoint

Parse via `httpx.ParseRequest`, call the use case, answer with `httpx.Write*Response` or `httpx.WriteError`.

**Rules:**

- ✅ A request DTO per endpoint with `validate:` tags (`required,min=1`).
- ✅ Setters answer `httpx.WriteCreatedResponse`, reads `httpx.WriteOKResponse`.
- ✅ Every error goes to `httpx.WriteError` (maps `customerr` type → HTTP status).
- ✅ Path params via `httpx.GetPathParam`.
- ❌ No business logic or repository access in the endpoint. ❌ Don't hand-roll JSON decode/validate.

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

	output, err := e.createTicket.CreateTicket(ctx, usecase.CreateTicketInput{
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

## 6. http/route.go

Register under `/api/v1/...` (the router is mounted at `/api`; group resource routes under `/v1/...`).

**Rules:**

- ✅ `r.Route("/v1/<resources>", ...)` grouping routes per aggregate.
- ✅ `RegisterRoutes` is passed the `*endpoint.Endpoint` built by `SetupEndpoints`.
- ❌ No usecase calls in the router — routing only.

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

## 7. Wiring — `<domain>/<domain>.go`

Composition root. Build the transaction manager, postgres repositories, usecases, endpoints; register routes.

**Rules:**

- ✅ One function `New(r chi.Router, db *sqlx.DB) error` per module.
- ✅ Ports are implemented by postgres/console adapters; no `mockery` mocks here.
- ❌ No wiring in endpoints/usecases — declarative, top-down.

```go
// internal/ticket/ticket.go
func New(r chi.Router, db *sqlx.DB) error {
	txManager := postgres.NewTransactionManager(db)

	ticketRepo, err := postgres.NewTicketRepository(db)
	if err != nil {
		return err
	}

	ep := http.SetupEndpoints(
		usecase.NewCreateTicketUsecase(ticketRepo),
		usecase.NewAssignTicketUsecase(txManager, console.NewNotifier()),
		usecase.NewGetTicketUsecase(ticketRepo),
		usecase.NewListTicketsUsecase(ticketRepo),
	)

	http.RegisterRoutes(r, ep)

	return nil
}
```

Read adapters for other domains go in `<domain>/public.go`:

```go
// internal/ticket/public.go
func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := postgres.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	getTicketUc := usecase.NewGetTicketUsecase(ticketRepo)

	return public.NewTicketReader(getTicketUc), nil
}
```

---

## Structural changes

Not covered here — see "Structure" in `AGENTS.md`, [docs/versioning.md](versioning.md), and [docs/cross-module-reads.md](cross-module-reads.md). This file only covers the fixed layer order for implementing a single domain feature.
