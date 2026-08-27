package endpoint

import (
	"net/http"
	"time"

	"github.com/charmingruby/lab/internal/shared/httpx"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

type TicketV1Response struct {
	AssigneeID  *string   `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
}

func newTicketV1Response(ticket *model.Ticket) TicketV1Response {
	return TicketV1Response{
		ID:          ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		Status:      string(ticket.Status),
		Priority:    string(ticket.Priority),
		AssigneeID:  ticket.AssigneeID,
		CreatedAt:   ticket.CreatedAt,
	}
}

func (e *Endpoint) GetTicketV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ticketID, err := httpx.GetPathParam(r, "id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	ticket, err := e.getTicket.GetTicket(ctx, usecase.GetTicketInput{
		TicketID: ticketID,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteOKResponse(w, newTicketV1Response(ticket))
}
