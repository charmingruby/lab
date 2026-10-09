package usecase

import (
	"context"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
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

func (u *Usecase) ListTickets(ctx context.Context, input ListTicketsInput) (ListTicketsOutput, error) {
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
