package usecase

import (
	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/internal/ticket/repository"
)

// Usecase orchestrates one business action per method. It is a concrete
// struct on purpose: there is a single implementation, so an interface
// would only serve mocks — and there are none.
type Usecase struct {
	ticketRepo *repository.TicketRepository
	txManager  *repository.TransactionManager
	notifier   client.NotificationClient
}

func New(
	ticketRepo *repository.TicketRepository,
	txManager *repository.TransactionManager,
	notifier client.NotificationClient,
) *Usecase {
	return &Usecase{
		ticketRepo: ticketRepo,
		txManager:  txManager,
		notifier:   notifier,
	}
}
