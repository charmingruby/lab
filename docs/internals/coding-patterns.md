# Coding Patterns

Implementation reference for a domain feature. Rules below are the pattern; code blocks are the current illustration (`internal/ticket/`, postgres, HTTP). Mirror the shape for any `<domain>`, do not invent new shapes.

> **Module structure**: see [architecture.md](./architecture.md).

## 1. Model

Constructor, invariants, state changes as explicit methods.

**Rules:**

- ✅ Constructor returns `(*Model, error)` when there are invariants to hold; plain `*Model` otherwise.
- ✅ Invalid state returns a package-level `var Err...`, not a string.
- ✅ State changes are methods named as verbs (`Assign`, `Resolve`).
- ✅ Legitimate values are typed constants with `Valid()`.
- ❌ Never expose raw field mutation from outside.

- Current: state changes use `core.Model.Touch`.

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

The golden source in `internal/<domain>/repository/`. No interface, no subpackage per provider.

**Rules:**

- ✅ One concrete struct per aggregate, methods named by business meaning.
- ✅ All methods take `context.Context` first and return the model or count.
- ✅ Not-found returns `(nil, nil)`.
- ❌ No interface for the golden source — usecases take the concrete repository.
- ❌ No inline queries in methods — always via prepared statements.

- Current: postgres via `sqlx`/`postgrex` — query map prepared once in the constructor, `deleted_at IS NULL` on reads, per-method timeout, `LIMIT/OFFSET` plus `COUNT(*)`. `sql.ErrNoRows` maps to `(nil, nil)`.

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

Transactions group the concrete repositories in a `Transaction` struct, run through `TransactionManager.Transact`:

```go
// internal/ticket/repository/transaction_manager.go
type Transaction struct {
	TicketRepo *TicketRepository
}
```

- Current: backed by `postgrex.RunInTx`.

---

## 3. usecase

One file per implemented use case, plus the concrete `Usecase` struct in `usecase/usecase.go`. It takes the concrete repository and transaction manager plus client ports.

**Rules:**

- ✅ Input/output structs live in the usecase file, named `<Verb><Resource>Input` / `<Verb><Resource>Output`.
- ✅ Infra failures wrap as integration errors.
- ✅ Domain outcomes map: missing → `NotFound`, duplicate → `Conflict`, invalid → `Validation`.
- ✅ Multi-repo writes go through the transaction manager.
- ✅ One `*Usecase` built with `New(...)`, shared by endpoints and public readers.
- ❌ No interface on the usecase — a single implementation needs no port.
- ❌ No transport types in the usecase. ❌ No raw storage handle — the concrete repository only.

- Current: `customerr.Validation / NotFound / Conflict / Integration`; transaction via `repository.TransactionManager.Transact`.

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

One delivery mechanism. Parse the request, call the use case, write the response.

**Rules:**

- ✅ A request DTO per endpoint with validation tags.
- ✅ Setters answer "created", reads answer "ok".
- ✅ Every error goes to the shared error writer (maps typed error → transport status).
- ✅ Path params via the shared helper.
- ❌ No business logic or repository access in the endpoint. ❌ Don't hand-roll decode/validate.

- Current: `httpx.ParseRequest`, `httpx.Write*Response` / `httpx.WriteError`, `httpx.GetPathParam`, `validate:"required,min=1"`, chi handler `func(w http.ResponseWriter, r *http.Request)`.

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

Register versioned routes under one registrar per mechanism.

**Rules:**

- ✅ Group routes per aggregate under `/v1/...`.
- ✅ Registrar receives the built endpoint; routing only.
- ❌ No usecase calls in the router.

- Current: router mounted at `/api`, chi `r.Route("/v1/<resources>")`, `RegisterRoutes(r, ep)`.

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

- ✅ One constructor per module.
- ✅ Concrete repository + client adapters wired here.
- ❌ No wiring in endpoints/usecases — declarative, top-down.

- Current: `New(r chi.Router, db *sqlx.DB) error`; adapters e.g. `console.NewNotifier()`.

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
func NewTicketReader(db *sqlx.DB) (public.Reader, error) {
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
