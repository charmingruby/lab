package endpoint

import (
	"net/http"

	"github.com/charmingruby/lab/internal/shared/httpx"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

type ListTicketsV1Response struct {
	Tickets    []TicketV1Response `json:"tickets"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int                `json:"total"`
	TotalPages int                `json:"total_pages"`
}

func (e *Endpoint) ListTicketsV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status, err := httpx.GetQueryParam(r, "status")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	params := httpx.GetPaginationParams(r)

	output, err := e.listTickets.ListTickets(ctx, usecase.ListTicketsInput{
		Status: status,
		Params: params,
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
		Tickets:    tickets,
		Page:       output.Page,
		Limit:      output.Limit,
		Total:      output.Total,
		TotalPages: output.TotalPages,
	})
}
