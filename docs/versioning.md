# Versioning

- **Unit of versioning: the endpoint file, not the folder.** A breaking change to `create_payment_v1.go` adds `create_payment_v2.go` beside it. The rest of `endpoint/` stays on v1.
- **`route.go` never carries a version suffix.** It's the single registrar mapping every endpoint version to its route (`/api/v1/...`, `/api/v2/...`). Deprecate `/v1` by removing its registration once callers migrate — don't mutate `/v1`'s behavior in place.
- **Non-HTTP protocols version the same way**, at the caller-facing boundary: gRPC registers `ServiceV1`/`ServiceV2`; a queue — a delivery mechanism in both directions — binds versioned events or queue names on the message boundary for producers and consumers alike.
- **Never version `model/`, `usecase/`, or `repository/`.** These stay single-version regardless of how many API versions call them.

## Example — v1 → v2

A breaking change (e.g. `charged_amount` becomes `amount`) adds a new endpoint file; the old one stays untouched.

**1. `http/endpoint/create_payment_v1.go`** — unchanged:

```go
type CreatePaymentV1Request struct {
	UserID        string `json:"user_id"        validate:"required,min=1"`
	OfferingID    string `json:"offering_id"    validate:"required,min=1"`
	ExternalID    string `json:"external_id"    validate:"required,min=1"`
	ChargedAmount int    `json:"charged_amount" validate:"required"`
}

func (e *Endpoint) CreatePaymentV1(w http.ResponseWriter, r *http.Request) {
	// ...calls e.createPayment.CreatePayment(...)
}
```

**2. `http/endpoint/create_payment_v2.go`** — new DTO + handler beside it, same use case:

```go
type CreatePaymentV2Request struct {
	UserID     string `json:"user_id"     validate:"required,min=1"`
	OfferingID string `json:"offering_id" validate:"required,min=1"`
	ExternalID string `json:"external_id" validate:"required,min=1"`
	Amount     int    `json:"amount"      validate:"required"`
}

func (e *Endpoint) CreatePaymentV2(w http.ResponseWriter, r *http.Request) {
	// ...maps Amount -> ChargedAmount, same e.createPayment.CreatePayment(...)
}
```

**3. `http/route.go`** — registers both until callers migrate, then drops v1:

```go
r.Route("/v1/payments", func(r chi.Router) {
	r.Post("/", ep.CreatePaymentV1) // deprecated: remove once callers are on v2
})

r.Route("/v2/payments", func(r chi.Router) {
	r.Post("/", ep.CreatePaymentV2)
})
```
