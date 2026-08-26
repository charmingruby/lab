package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/internal/ticket/model"
	"github.com/charmingruby/new/internal/ticket/repository"
)

type CreateTicketInput = model.TicketInput

type CreateTicketOutput struct {
	ID string
}

type createTicketUsecase struct {
	ticketRepo repository.TicketRepository
}

func NewCreateTicketUsecase(ticketRepo repository.TicketRepository) *createTicketUsecase {
	return &createTicketUsecase{
		ticketRepo: ticketRepo,
	}
}

func (u *createTicketUsecase) CreateTicket(
	ctx context.Context,
	input CreateTicketInput,
) (CreateTicketOutput, error) {
	ticket, err := model.NewTicket(input)
	if err != nil {
		return CreateTicketOutput{}, customerr.Validation(err.Error())
	}

	if err := u.ticketRepo.Create(ctx, ticket); err != nil {
		return CreateTicketOutput{}, customerr.Integration(err)
	}

	return CreateTicketOutput{
		ID: ticket.ID,
	}, nil
}
