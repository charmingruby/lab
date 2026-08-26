package public

import (
	"context"

	"github.com/charmingruby/new/internal/ticket/usecase"
)

type TicketReader struct {
	getTicket usecase.GetTicketUsecase
}

func NewTicketReader(getTicket usecase.GetTicketUsecase) *TicketReader {
	return &TicketReader{getTicket: getTicket}
}

func (r *TicketReader) GetTicketStatus(ctx context.Context, ticketID string) (string, error) {
	t, err := r.getTicket.GetTicket(ctx, usecase.GetTicketInput{
		TicketID: ticketID,
	})
	if err != nil {
		return "", err
	}

	return string(t.Status), nil
}
