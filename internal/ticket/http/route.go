package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/new/internal/ticket/http/endpoint"
	"github.com/charmingruby/new/internal/ticket/usecase"
)

func SetupEndpoints(
	createTicket usecase.CreateTicketUsecase,
	assignTicket usecase.AssignTicketUsecase,
	getTicket usecase.GetTicketUsecase,
	listTickets usecase.ListTicketsUsecase,
) *endpoint.Endpoint {
	return endpoint.New(
		createTicket,
		assignTicket,
		getTicket,
		listTickets,
	)
}

func RegisterRoutes(
	r chi.Router,
	ep *endpoint.Endpoint,
) {
	r.Route("/v1/tickets", func(r chi.Router) {
		r.Post("/", ep.CreateTicketV1)
		r.Get("/", ep.ListTicketsV1)
		r.Get("/{id}", ep.GetTicketV1)
		r.Patch("/{id}/assign", ep.AssignTicketV1)
	})
}
