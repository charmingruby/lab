package public

import (
	"context"

	"github.com/charmingruby/lab/internal/ticket/usecase"
)

type Reader interface {
	GetTicketStatus(ctx context.Context, ticketID string) (string, error)
}

type TicketReader struct {
	uc *usecase.Usecase
}

func NewTicketReader(uc *usecase.Usecase) Reader {
	return &TicketReader{uc: uc}
}

func (r *TicketReader) GetTicketStatus(ctx context.Context, ticketID string) (string, error) {
	t, err := r.uc.GetTicket(ctx, usecase.GetTicketInput{
		TicketID: ticketID,
	})
	if err != nil {
		return "", err
	}

	return string(t.Status), nil
}
