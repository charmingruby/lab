package usecase

import (
	"context"

	"github.com/charmingruby/lab/internal/ticket/model"
)

type CreateTicketUsecase interface {
	CreateTicket(ctx context.Context, input CreateTicketInput) (CreateTicketOutput, error)
}

type AssignTicketUsecase interface {
	AssignTicket(ctx context.Context, input AssignTicketInput) error
}

type GetTicketUsecase interface {
	GetTicket(ctx context.Context, input GetTicketInput) (*model.Ticket, error)
}

type ListTicketsUsecase interface {
	ListTickets(ctx context.Context, input ListTicketsInput) (ListTicketsOutput, error)
}
