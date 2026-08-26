package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/shared/httpx"
	"github.com/charmingruby/new/internal/ticket/usecase"
)

type CreateTicketV1Request struct {
	Title       string `json:"title"       validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	Priority    string `json:"priority"    validate:"required,min=1"`
}

type CreateTicketV1Response struct {
	ID string `json:"id"`
}

func (e *Endpoint) CreateTicketV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreateTicketV1Request](w, r)
	if err != nil {
		return
	}

	output, err := e.createTicket.CreateTicket(ctx, usecase.CreateTicketInput{
		Title:       request.Title,
		Description: request.Description,
		Priority:    request.Priority,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreateTicketV1Response{
		ID: output.ID,
	})
}
