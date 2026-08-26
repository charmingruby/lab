package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/internal/ticket/model"
	"github.com/charmingruby/new/internal/ticket/repository"
)

type ListTicketsInput struct {
	Status string
	Page   int
}

type ListTicketsOutput struct {
	Tickets []model.Ticket
	Total   int
}

type listTicketsUsecase struct {
	ticketRepo repository.TicketRepository
}

func NewListTicketsUsecase(ticketRepo repository.TicketRepository) *listTicketsUsecase {
	return &listTicketsUsecase{
		ticketRepo: ticketRepo,
	}
}

func (u *listTicketsUsecase) ListTickets(ctx context.Context, input ListTicketsInput) (ListTicketsOutput, error) {
	params := core.DefaultPaginationParams(input.Page)

	tickets, total, err := u.ticketRepo.ListByStatus(ctx, input.Status, params)
	if err != nil {
		return ListTicketsOutput{}, customerr.Integration(err)
	}

	return ListTicketsOutput{
		Tickets: tickets,
		Total:   total,
	}, nil
}
