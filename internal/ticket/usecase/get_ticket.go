package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/internal/ticket/model"
	"github.com/charmingruby/new/internal/ticket/repository"
)

type GetTicketInput struct {
	TicketID string
}

type getTicketUsecase struct {
	ticketRepo repository.TicketRepository
}

func NewGetTicketUsecase(ticketRepo repository.TicketRepository) *getTicketUsecase {
	return &getTicketUsecase{
		ticketRepo: ticketRepo,
	}
}

func (u *getTicketUsecase) GetTicket(ctx context.Context, input GetTicketInput) (*model.Ticket, error) {
	ticket, err := u.ticketRepo.FindByID(ctx, input.TicketID)
	if err != nil {
		return nil, customerr.Integration(err)
	}

	if ticket == nil {
		return nil, customerr.NotFound("ticket not found")
	}

	return ticket, nil
}
