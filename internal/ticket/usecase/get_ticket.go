package usecase

import (
	"context"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
)

type GetTicketInput struct {
	TicketID string
}

func (u *Service) GetTicket(ctx context.Context, input GetTicketInput) (*model.Ticket, error) {
	ticket, err := u.ticketRepo.FindByID(ctx, input.TicketID)
	if err != nil {
		return nil, customerr.Integration(err)
	}

	if ticket == nil {
		return nil, customerr.NotFound("ticket not found")
	}

	return ticket, nil
}
