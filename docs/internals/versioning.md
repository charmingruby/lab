# Versioning

Same shape for any `<domain>` and any delivery mechanism — current illustration is HTTP.

- **Unit of versioning: the endpoint, not the folder.** A breaking change adds a new endpoint beside the old one. The rest stays on the old version.
- **The registrar never carries a version suffix.** It's the single place mapping every endpoint version to its route. Deprecate the old version by removing its registration once callers migrate — don't mutate its behavior in place.
- **Every protocol versions at the caller-facing boundary**: HTTP registers a new route, gRPC a new service, messaging a new event/queue version for producers and consumers alike.
- **Never version `model/`, `usecase/`, or `repository/`.** These stay single-version regardless of how many API versions call them.

- Current: `http/endpoint/create_ticket_v1.go` → `create_ticket_v2.go`, `http/route.go` maps `/v1/...` → `/v2/...`.

Source: [internal/ticket/http/endpoint/](../../internal/ticket/http/endpoint/), [internal/ticket/http/route.go](../../internal/ticket/http/route.go)

## Example — v1 → v2

A breaking change (e.g. `title` becomes `name`) adds a new endpoint file; the old one stays untouched.

**1. [internal/ticket/http/endpoint/create_ticket_v1.go](../../internal/ticket/http/endpoint/create_ticket_v1.go)** — unchanged (current: chi + `httpx`):

```go
type CreateTicketV1Request struct {
	Title       string `json:"title"       validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	Priority    string `json:"priority"    validate:"required,min=1"`
}

func (e *Endpoint) CreateTicketV1(w http.ResponseWriter, r *http.Request) {
	// ...calls e.uc.CreateTicket(...)
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
	// ...maps Name -> Title, same e.uc.CreateTicket(...)
}
```

**3. [internal/ticket/http/route.go](../../internal/ticket/http/route.go)** — registers both until callers migrate, then drops v1 (current: chi):

```go
r.Route("/v1/tickets", func(r chi.Router) {
	r.Post("/", ep.CreateTicketV1) // deprecated: remove once callers are on v2
})

r.Route("/v2/tickets", func(r chi.Router) {
	r.Post("/", ep.CreateTicketV2)
})
```
