package usecase

import (
	"context"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/repository"
)

type Usecase interface {
	CreateTicket(ctx context.Context, input CreateTicketInput) (CreateTicketOutput, error)
	AssignTicket(ctx context.Context, input AssignTicketInput) error
	GetTicket(ctx context.Context, input GetTicketInput) (*model.Ticket, error)
	ListTickets(ctx context.Context, input ListTicketsInput) (ListTicketsOutput, error)
}

type Service struct {
	ticketRepo repository.TicketRepository
	txManager  core.TransactionManager[repository.Transaction]
	notifier   client.NotificationClient
}

func New(
	ticketRepo repository.TicketRepository,
	txManager core.TransactionManager[repository.Transaction],
	notifier client.NotificationClient,
) *Service {
	return &Service{
		ticketRepo: ticketRepo,
		txManager:  txManager,
		notifier:   notifier,
	}
}
