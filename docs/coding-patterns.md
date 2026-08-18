# Coding Patterns

Implementation reference for a domain feature. Real code from `internal/billing/`. Mirror it — do not invent new shapes.

> **Module structure**: follow "Structure" in `AGENTS.md`. For domain-modeling, tests, or versioning, see the "Progressive Documentation Loading" table.

## The Dependency Spine

A domain is a ports-and-adapters module in `internal/<domain>`:

```
<protocol> → usecase → repository (port) → repository/postgres
                              → client (port)   → client/console
                              → model
```

**Rules:**

- ✅ One layer per responsibility: `model` (state), `repository` (port), `repository/postgres` (SQL), `usecase` (business), `http` (transport).
- ✅ Wire everything declaratively in `<domain>/<domain>.go`.
- ❌ Never add a service layer or any abstraction between usecase and repository.
- ❌ Never let an endpoint call a repository or client directly — go through the usecase.
- ❌ Never import another module's `usecase`/`repository`/`model` — use its `client` port.

---

## 1. Model

Constructor (`New<Model>(input)`), invariants, and state changes as explicit methods using `core.Model.Touch`.

**Rules:**

- ✅ Constructor returns `(*Model, error)` when there are invariants to hold; plain `*Model` otherwise.
- ✅ Invalid state returns a package-level `var Err...`, not a string.
- ✅ State changes are methods named as verbs (`Activate`, `MarkAsPaid`), using `Touch(func(m *core.Model))`.
- ✅ Legitimate values are typed constants with `Valid()`.
- ❌ Never expose raw field mutation from outside — fields are embedded in the model, changed via methods.
- ❌ Never assign `Status = "paid"` strings outside the model package.

```go
// internal/billing/model/offering.go
func NewOffering(input OfferingInput) (*Offering, error) {
	chargeType := ChargeType(input.ChargeType)

	if !chargeType.Valid() {
		return nil, ErrInvalidChargeType
	}

	return &Offering{
		Model:       core.NewModel(),
		Name:        input.Name,
		Description: input.Description,
		ChargeType:  chargeType,
		Currency:    input.Currency,
		Price:       input.Price,
		IsActive:    input.IsActive,
	}, nil
}

func (o *Offering) Activate() {
	o.Touch(func(m *core.Model) {
		o.IsActive = true
	})
}
```

BAD — endpoint changing state directly:

```go
// ❌ In an endpoint: reaching into the fields
payment.Status = "paid"   // ❌ string literal + no UpdateAt touch

// ✅ In the usecase: explicit method
payment.MarkAsPaid()      // sets status + touches UpdatedAt
```

---

## 2. Repository (port)

Interface at the root: `internal/billing/repository/repository.go`.

**Rules:**

- ✅ One interface per aggregate, methods named by business meaning.
- ✅ All methods take `context.Context` first and return the model or count.
- ✅ Not-found contracts are documented by the postgres impl: returns `(nil, nil)`, not `sql.ErrNoRows`.
- ❌ No ORM/SQL types in the port (no `sqlx`), so usecases and mocks stay transport-free.

```go
// internal/billing/repository/repository.go
type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	FindByID(ctx context.Context, id string) (*model.Payment, error)
	FindByExternalID(ctx context.Context, externalID string) (*model.Payment, error)
	ListByUserID(ctx context.Context, userID string, params core.PaginationParams) ([]model.Payment, int, error)
	Update(ctx context.Context, payment *model.Payment) error
}
```

---

## 3. repository/postgres

Prepared statements via `postgrex.Querier`.

**Rules:**

- ✅ Query map as `var paymentQueries = map[string]string{ ... }`, prepared once in the constructor.
- ✅ `deleted_at IS NULL` on every read.
- ✅ `context.WithTimeout(ctx, postgrex.DefaultReadTimeout)` per method.
- ✅ Not-found returns `(nil, nil)` — never surface `sql.ErrNoRows` to callers.
- ✅ `LIMIT $2 OFFSET $3` pagination plus a separate `COUNT(*)` query.
- ❌ No inline queries in methods — always via the prepared `statement(name)`.
- ❌ Never return `sql.ErrNoRows`.

```go
// internal/billing/repository/postgres/payment_repository.go
func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*model.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findPaymentByIDQuery)
	if err != nil {
		return nil, err
	}

	var payment model.Payment
	if err := stmt.QueryRowxContext(ctx, id).StructScan(&payment); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // ✅ no row -> nil, nil
		}
		return nil, err
	}

	return &payment, nil
}
```

BAD — leaking `sql.ErrNoRows`:

```go
// ❌ Callers would have to import database/sql
if errors.Is(err, sql.ErrNoRows) {
	return nil, sql.ErrNoRows   // ❌
}
```

---

## 4. usecase

One file per implemented use case; interfaces aggregated in `usecase/usecase.go` (what mockery mocks).

**Rules:**

- ✅ Input/output structs live in the usecase file, named `<Verb><Resource>Input` / `<Verb><Resource>Output` (e.g. `CreatePaymentInput`, `GetPaymentInput`, `ListPaymentsOutput`).
- ✅ Infra failures wrap as `customerr.Integration(err)`.
- ✅ Domain outcomes map: missing → `customerr.NotFound`, duplicate → `customerr.Conflict`, invalid → `customerr.Validation`.
- ✅ Multi-repo writes go through `core.TransactionManager[repository.Transaction]` (`postgrex.RunInTx`).
- ✅ Interface stays in `usecase.go`, typed by the concrete `*usecase` constructing it.
- ❌ No HTTP/`net/http` types in the usecase.
- ❌ No `*sqlx.DB` in the usecase — repositories only.

```go
// internal/billing/usecase/get_payment.go
func (u *getPaymentUsecase) GetPayment(ctx context.Context, input GetPaymentInput) (*model.Payment, error) {
	payment, err := u.paymentRepo.FindByID(ctx, input.PaymentID)
	if err != nil {
		return nil, customerr.Integration(err)
	}

	if payment == nil {
		return nil, customerr.NotFound("payment not found")
	}

	return payment, nil
}
```

Multi-repo write, transactional (`create_payment.go`):

```go
// internal/billing/usecase/create_payment.go
err = u.txManager.Transact(func(tx repository.Transaction) error {
	existing, err := tx.PaymentRepo.FindByExternalID(ctx, input.ExternalID)
	if err != nil {
		return customerr.Integration(err)
	}

	if existing != nil {
		paymentID = existing.ID
		return nil
	}

	payment := model.NewPayment(model.PaymentInput{ /* ... */ })
	payment.MarkAsPaid()

	if err := tx.PaymentRepo.Create(ctx, payment); err != nil {
		return customerr.Integration(err)
	}

	paymentID = payment.ID
	return nil
})
```

The `Transaction` bundle (sets of repos usable inside a tx) is declared in `repository/repository.go`:

```go
type Transaction struct {
	PaymentRepo  PaymentRepository
	OfferingRepo OfferingRepository
}
```

The postgres manager wires it to `postgrex.RunInTx` (`repository/postgres/transaction_manager.go`).

BAD — skipping the transaction for two writes, or calling the repo twice (once for check, once inside tx) — use the tx copy:

```go
// ❌ Non-transactional read+write
found, _ := u.paymentRepo.FindByExternalID(ctx, input.ExternalID)
...
u.paymentRepo.Create(ctx, payment) // ❌ not atomic with the check

// ✅ Transactional: both reads and writes inside Transact
u.txManager.Transact(func(tx repository.Transaction) error { ... })
```

---

## 5. http/endpoint

Parse via `httpx.ParseRequest`, call the usecase, answer with `httpx.Write*Response` or `httpx.WriteError`.

**Rules:**

- ✅ A request DTO per endpoint with `validate:` tags (`required,min=1`).
- ✅ Setters answer `httpx.WriteCreatedResponse`, reads `httpx.WriteOKResponse`.
- ✅ Every error goes to `httpx.WriteError` (maps `customerr` type → HTTP status).
- ✅ Path params via `httpx.GetPathParam`.
- ✅ Method is a handler on the `Endpoint` struct: `func (e *Endpoint) CreatePaymentV1(w, r)`.
- ❌ No business logic or repository access in the endpoint.
- ❌ Don't hand-roll JSON decode/validate — use `httpx.ParseRequest`.

```go
// internal/billing/http/endpoint/create_payment_v1.go
type CreatePaymentV1Request struct {
	UserID        string `json:"user_id"        validate:"required,min=1"`
	OfferingID    string `json:"offering_id"    validate:"required,min=1"`
	ExternalID    string `json:"external_id"    validate:"required,min=1"`
	ChargedAmount int    `json:"charged_amount" validate:"required"`
}

func (e *Endpoint) CreatePaymentV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreatePaymentV1Request](w, r)
	if err != nil {
		return
	}

	output, err := e.createPayment.CreatePayment(ctx, usecase.CreatePaymentInput{
		UserID:        request.UserID,
		OfferingID:    request.OfferingID,
		ExternalID:    request.ExternalID,
		ChargedAmount: request.ChargedAmount,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreatePaymentV1Response{
		ID: output.ID,
	})
}
```

BAD — fat handler:

```go
// ❌ Endpoint reaching into the repository + raw JSON
var req struct{ ... }
json.NewDecoder(r.Body).Decode(&req)
payment := &model.Payment{ /* field assignments */ } // ❌ no NewPayment, no invariants
db.ExecContext(ctx, "INSERT INTO payments ...")       // ❌ SQL in the handler
```

---

## 6. http/route.go

Register under `/api/v1/...` (the router is mounted at `/api`; group resource routes under `/v1/...` to land on `/api/v1/...`).

**Rules:**

- ✅ `r.Route("/v1/<resources>", ...)` grouping routes per aggregate.
- ✅ `RegisterRoutes` is passed the `*endpoint.Endpoint` built by `SetupEndpoints`.
- ❌ No usecase calls in the router — routing only.

```go
// internal/billing/http/route.go
func RegisterRoutes(r chi.Router, ep *endpoint.Endpoint) {
	r.Route("/v1/offerings", func(r chi.Router) {
		r.Post("/", ep.CreateOfferingV1)
	})

	r.Route("/v1/payments", func(r chi.Router) {
		r.Post("/", ep.CreatePaymentV1)
		r.Get("/", ep.ListPaymentsV1)
		r.Get("/{id}", ep.GetPaymentV1)
	})
}
```

---

## 7. Wiring the module — `<domain>/<domain>.go`

Composition root. Build the transaction manager, postgres repositories, usecases, endpoints; register routes.

**Rules:**

- ✅ One function `New(r chi.Router, db *sqlx.DB) error` per module.
- ✅ Ports are implemented by postgres/console adapters; no `mockery` mocks here.
- ❌ No wiring in endpoints/usecases — declarative, top-down.

```go
// internal/billing/billing.go
func New(r chi.Router, db *sqlx.DB) error {
	txManager := postgres.NewTransactionManager(db)

	offeringRepo, err := postgres.NewOfferingRepository(db)
	if err != nil {
		return err
	}
	paymentRepo, err := postgres.NewPaymentRepository(db)
	if err != nil {
		return err
	}

	ep := http.SetupEndpoints(
		usecase.NewCreateOfferingUsecase(offeringRepo),
		usecase.NewCreatePaymentUsecase(txManager, offeringRepo, paymentRepo, console.NewNotifier()),
		usecase.NewGetPaymentUsecase(paymentRepo),
		usecase.NewListPaymentsUsecase(paymentRepo),
	)

	http.RegisterRoutes(r, ep)

	return nil
}
```

Read adapters for other domains go in `<domain>/public.go`, using the domain's own `client` faces:

```go
// internal/billing/public.go
func NewPaymentReader(db *sqlx.DB) (client.PaymentReader, error) {
	paymentRepo, err := postgres.NewPaymentRepository(db)
	if err != nil {
		return nil, err
	}

	return public.NewPaymentReader(usecase.NewGetPaymentUsecase(paymentRepo)), nil
}
```

---

## Structural changes (new protocol, new port backend, cross-module read, external dependency)

Not covered here — see "Structure" in `AGENTS.md`, [docs/versioning.md](versioning.md), and [docs/cross-module-reads.md](cross-module-reads.md). This file only covers the fixed layer order for implementing a single domain feature.
