package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/shared/httpx"
	"github.com/charmingruby/new/internal/ticket/usecase"
)

type AssignTicketV1Request struct {
	AssigneeID string `json:"assignee_id" validate:"required,min=1"`
}

func (e *Endpoint) AssignTicketV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ticketID, err := httpx.GetPathParam(r, "id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	request, err := httpx.ParseRequest[AssignTicketV1Request](w, r)
	if err != nil {
		return
	}

	err = e.assignTicket.AssignTicket(ctx, usecase.AssignTicketInput{
		TicketID:   ticketID,
		AssigneeID: request.AssigneeID,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteEmptyResponse(w)
}
