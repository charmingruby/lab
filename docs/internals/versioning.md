# Versioning

- **Unit of versioning: the endpoint file, not the folder.** A breaking change to `create_ticket_v1.go` adds `create_ticket_v2.go` beside it. The rest of `endpoint/` stays on v1.
- **`route.go` never carries a version suffix.** It's the single registrar mapping every endpoint version to its route (`/api/v1/...`, `/api/v2/...`). Deprecate `/v1` by removing its registration once callers migrate — don't mutate `/v1`'s behavior in place.
- **Non-HTTP protocols version the same way**, at the caller-facing boundary: gRPC registers `ServiceV1`/`ServiceV2`; a queue — a delivery mechanism in both directions — binds versioned events or queue names on the message boundary for producers and consumers alike.
- **Never version `model/`, `usecase/`, or `repository/`.** These stay single-version regardless of how many API versions call them.

Source: [internal/ticket/http/endpoint/](../../internal/ticket/http/endpoint/), [internal/ticket/http/route.go](../../internal/ticket/http/route.go)

## Example — v1 → v2

A breaking change (e.g. `title` becomes `name`) adds a new endpoint file; the old one stays untouched.

**1. [internal/ticket/http/endpoint/create_ticket_v1.go](../../internal/ticket/http/endpoint/create_ticket_v1.go)** — unchanged:

```go
type CreateTicketV1Request struct {
	Title       string `json:"title"       validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	Priority    string `json:"priority"    validate:"required,min=1"`
}

func (e *Endpoint) CreateTicketV1(w http.ResponseWriter, r *http.Request) {
	// ...calls e.createTicket.CreateTicket(...)
}
```

**2. `http/endpoint/create_ticket_v2.go`** — new DTO + handler beside it, same use case:

```go
type CreateTicketV2Request struct {
	Name        string `json:"name"        validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	Priority    string `json:"priority"    validate:"required,min=1"`
}

func (e *Endpoint) CreateTicketV2(w http.ResponseWriter, r *http.Request) {
	// ...maps Name -> Title, same e.createTicket.CreateTicket(...)
}
```

**3. [internal/ticket/http/route.go](../../internal/ticket/http/route.go)** — registers both until callers migrate, then drops v1:

```go
r.Route("/v1/tickets", func(r chi.Router) {
	r.Post("/", ep.CreateTicketV1) // deprecated: remove once callers are on v2
})

r.Route("/v2/tickets", func(r chi.Router) {
	r.Post("/", ep.CreateTicketV2)
})
```
