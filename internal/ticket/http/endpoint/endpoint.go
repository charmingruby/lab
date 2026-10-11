package endpoint

import (
	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/lab/internal/ticket/usecase"
)

type Endpoint struct {
	uc *usecase.Usecase
}

func New(r chi.Router, uc *usecase.Usecase) *Endpoint {
	ep := &Endpoint{
		uc: uc,
	}

	r.Route("/v1/tickets", func(r chi.Router) {
		r.Post("/", ep.CreateTicketV1)
		r.Get("/{id}", ep.GetTicketV1)
		r.Patch("/{id}/assign", ep.AssignTicketV1)
	})

	return ep
}
