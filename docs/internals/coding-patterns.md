# Coding Patterns

Implementation reference for a domain feature. Rules below are the pattern; code blocks illustrate the shape for any `<domain>`. Mirror the shape, do not invent new shapes.

> **Module structure**: see [architecture.md](./architecture.md).

## 1. Model

Constructor, invariants, state changes as explicit methods.

**Rules:**

- ✅ Constructor returns `(*Model, error)` when there are invariants to hold; plain `*Model` otherwise.
- ✅ Invalid state returns a package-level `var Err...`, not a string.
- ✅ State changes are methods named as verbs (`Assign`, `Resolve`).
- ✅ Legitimate values are typed constants with `Valid()`.
- ✅ Fields may be mutated directly, but only inside `core.Model.Touch` so `UpdatedAt` stays correct.

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
- ✅ The repository imports exactly one `internal/platform/store/<provider>` toolkit; the golden source stays concrete.
- ✅ Query text lives in named constants grouped in one `map[string]string`, prepared once by the constructor.
- ✅ The constructor takes the store's querier (for SQL, `postgrex.Querier`), so the same repository works over a pool or a transaction.
- ✅ Methods delegate to the store helpers (for SQL, `postgrex.One[T]`, `Many[T]`, `Exec`, `ExecReturning`).
- ✅ All methods take `context.Context` first and return the model or count.
- ✅ Not-found returns `(nil, nil)`.
- ❌ No interface for the golden source — usecases take the concrete repository.
- ❌ No inline queries in methods — always a prepared statement looked up by name.
- ❌ No per-method timeout boilerplate — the helpers already bound reads and writes.
- ❌ No repository interface for a second database — a new system of record means a new `internal/platform/store/<provider>` toolkit plus a new concrete repository.

```go
// internal/<domain>/repository/<aggregate>_repository.go
const (
	find<Aggregate>ByIDQuery = "find <aggregate> by id"
	update<Aggregate>Query   = "update <aggregate>"
)

var <aggregate>Queries = map[string]string{
	find<Aggregate>ByIDQuery: `
		SELECT * FROM <aggregates>
		WHERE
			id=$1 AND
			deleted_at IS NULL`,
	update<Aggregate>Query: `
		UPDATE <aggregates>
		SET
			name=$1,
			updated_at=$2
		WHERE
			id=$3 AND
			deleted_at IS NULL`,
}

type <Aggregate>Repository struct {
	stmts *postgrex.Statements
}

func New<Aggregate>Repository(db postgrex.Querier) (*<Aggregate>Repository, error) {
	stmts, err := postgrex.Prepare(db, <aggregate>Queries)
	if err != nil {
		return nil, err
	}

	return &<Aggregate>Repository{stmts: stmts}, nil
}

func (r *<Aggregate>Repository) FindByID(ctx context.Context, id string) (*model.<Aggregate>, error) {
	return postgrex.One[model.<Aggregate>](ctx, r.stmts, find<Aggregate>ByIDQuery, id)
}

func (r *<Aggregate>Repository) Update(ctx context.Context, aggregate *model.<Aggregate>) error {
	return postgrex.Exec(ctx, r.stmts, update<Aggregate>Query,
		aggregate.Name,
		aggregate.UpdatedAt,
		aggregate.ID,
	)
}
```

Helper semantics:

| Helper             | Use                                       |
| ------------------ | ----------------------------------------- |
| `One[T]`           | single row → `*T`; no rows → `(nil, nil)` |
| `Many[T]`          | list of rows → `[]T` (never nil)          |
| `Exec`             | write with no return                      |
| `ExecReturning[T]` | write returning a single scalar           |

Reads and writes get a bounded context inside the helper (`DefaultReadTimeout` / `DefaultWriteTimeout`). A missing statement name surfaces as `postgrex.ErrQueryPreparation`.

Repositories are tied to the store they use. There is no `store` port: a second database is a second `internal/platform/store/<provider>` toolkit plus a second concrete repository, not an adapter swap. `Statements` and the query helpers are SQL-specific; a non-SQL store exposes its own primitives.

Transactions group the concrete repositories in a `Transaction` struct. `Transact` opens the transaction with `postgrex.RunInTx` and rebuilds each repository over the `*sqlx.Tx`:

```go
// internal/<domain>/repository/transaction_manager.go
type Transaction struct {
	<Aggregate>Repository *<Aggregate>Repository
}

func (t *TransactionManager) Transact(ctx context.Context, fn func(Transaction) error) error {
	return postgrex.RunInTx(ctx, t.db, func(tx *sqlx.Tx) error {
		aggregateRepo, err := New<Aggregate>Repository(tx)
		if err != nil {
			return err
		}

		return fn(Transaction{
			<Aggregate>Repository: aggregateRepo,
		})
	})
}
```

---

## 3. usecase

One file per implemented use case, plus the concrete `Usecase` struct in `usecase/usecase.go`. It takes the concrete repository and transaction manager plus client ports.

**Rules:**

- ✅ Input/output structs live in the usecase file, named `<Verb><Resource>Input` / `<Verb><Resource>Output`.
- ✅ Infra failures wrap as internal errors with `customerr.Internal(message, err)`.
- ✅ Domain outcomes map: missing → `NotFound`, duplicate → `Conflict`, invalid → `Validation`.
- ✅ Atomic writes (multi-repo or read-modify-write) go through the transaction manager. See [architecture.md](./architecture.md#transaction-safety).
- ✅ One `*Usecase` built with `New(...)`, shared by endpoints and public readers.
- ✅ Best-effort side effects (e.g. notifier) stay outside the transaction and never fail the usecase — log and continue. Best-effort describes failure handling, not asynchronous execution. If you write a test for the failure path, use the `memory.go` fake, see [testing.md](./testing.md).
- ❌ No interface on the usecase — a single implementation needs no port.
- ✅ Endpoints/listeners map inbound DTOs to usecase inputs. Usecases may construct outbound events using contract types defined outside the usecase package; event contracts are never defined in the usecase package.
- ❌ No raw storage handle — the concrete repository only.

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
		return CreateTicketOutput{}, customerr.Internal("failed to create ticket", err)
	}

	return CreateTicketOutput{ID: ticket.ID}, nil
}
```

Transactional case — `assign_ticket.go` (shape only; the real file also sends a best-effort notification after commit, which never fails the usecase):

```go
func (u *Usecase) AssignTicket(
	ctx context.Context,
	input AssignTicketInput,
) error {
	err := u.txManager.Transact(ctx, func(tx repository.Transaction) error {
		ticket, err := tx.TicketRepo.FindByID(ctx, input.TicketID)
		if err != nil {
			return customerr.Internal("failed to find ticket", err)
		}

		if ticket == nil {
			return customerr.NotFound("ticket not found")
		}

		if err := ticket.Assign(input.AssigneeID); err != nil {
			return customerr.Validation(err.Error())
		}

		if err := tx.TicketRepo.Update(ctx, ticket); err != nil {
			return customerr.Internal("failed to update ticket", err)
		}

		return nil
	})

	return err
}
```

---

## 4. http/endpoint

One delivery mechanism. Parse the request, call the use case, write the response. Route registration lives here too, in `New` — paths, verbs, and per-route middlewares stay next to the handlers they mount.

**Rules:**

- ✅ A request DTO per endpoint with validation tags.
- ✅ Setters answer "created", reads answer "ok".
- ✅ Every error goes to the shared error writer (typed errors map to statuses per [architecture.md](./architecture.md#error-flow); missing path params use `httpx.WriteBadRequest`).
- ✅ Path params via the shared helper.
- ✅ Group routes per aggregate under `/v1/...` in `New(r chi.Router, uc *usecase.Usecase)`.
- ❌ No business logic or repository access in the endpoint. ❌ Don't hand-roll decode/validate.

Source: [internal/ticket/http/endpoint/endpoint.go](../../internal/ticket/http/endpoint/endpoint.go), [internal/ticket/http/endpoint/create_ticket_v1.go](../../internal/ticket/http/endpoint/create_ticket_v1.go)

```go
// internal/ticket/http/endpoint/endpoint.go
func New(r chi.Router, uc *usecase.Usecase) *Endpoint {
	ep := &Endpoint{uc: uc}

	r.Route("/v1/tickets", func(r chi.Router) {
		r.Post("/", ep.CreateTicketV1)
		r.Get("/{id}", ep.GetTicketV1)
		r.Patch("/{id}/assign", ep.AssignTicketV1)
	})

	return ep
}
```

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
		httpx.HandleErr(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreateTicketV1Response{ID: output.ID})
}
```

---

## 5. Wiring — `<domain>/<domain>.go`

Composition root. Build the transaction manager, repositories, usecases; hand the router to the endpoint, which registers its own routes.

**Rules:**

- ✅ One composition root per module, with constructors for its transports and public readers as needed.
- ✅ Concrete repository + client adapters wired here.
- ❌ No wiring in usecases — declarative, top-down.

Source: [internal/ticket/ticket.go](../../internal/ticket/ticket.go)

```go
// internal/ticket/ticket.go
func New(r chi.Router, db *sqlx.DB) error {
	txManager := repository.NewTransactionManager(db)

	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return err
	}

	endpoint.New(r, usecase.New(ticketRepo, txManager, notifier.NewConsole()))

	return nil
}
```

Read adapters for other domains are assembled in the same file — shape owned by [cross-module-reads.md](./cross-module-reads.md), not duplicated here.

---

## Structural changes

Not covered here — see [architecture.md](./architecture.md), [versioning.md](versioning.md), and [cross-module-reads.md](cross-module-reads.md). This file only covers the fixed layer order for implementing a single domain feature.
