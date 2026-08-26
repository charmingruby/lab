package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/shared/httpx"
	"github.com/charmingruby/new/internal/ticket/usecase"
)

type ListTicketsV1Response struct {
	Tickets []TicketV1Response `json:"tickets"`
	Total   int                `json:"total"`
}

func (e *Endpoint) ListTicketsV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status, err := httpx.GetQueryParam(r, "status")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	page := httpx.GetPageQueryParam(r)

	output, err := e.listTickets.ListTickets(ctx, usecase.ListTicketsInput{
		Status: status,
		Page:   page,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	tickets := make([]TicketV1Response, 0, len(output.Tickets))
	for _, t := range output.Tickets {
		tickets = append(tickets, newTicketV1Response(&t))
	}

	httpx.WriteOKResponse(w, ListTicketsV1Response{
		Tickets: tickets,
		Total:   output.Total,
	})
}
