package usecase

import (
	"context"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/repository"
)

type ListTicketsInput struct {
	Status string
	Params core.PaginationParams
}

type ListTicketsOutput struct {
	Tickets    []model.Ticket
	Page       int
	Limit      int
	Total      int
	TotalPages int
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
	params := input.Params.Validate()

	tickets, total, err := u.ticketRepo.ListByStatus(ctx, input.Status, params)
	if err != nil {
		return ListTicketsOutput{}, customerr.Integration(err)
	}

	return ListTicketsOutput{
		Tickets:    tickets,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: params.TotalPages(total),
	}, nil
}
