package usecase

import (
	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/internal/ticket/repository"
)

type Usecase struct {
	ticketRepo *repository.TicketRepository
	txManager  *repository.TransactionManager
	notifier   client.Notifier
}

func New(
	ticketRepo *repository.TicketRepository,
	txManager *repository.TransactionManager,
	notifier client.Notifier,
) *Usecase {
	return &Usecase{
		ticketRepo: ticketRepo,
		txManager:  txManager,
		notifier:   notifier,
	}
}
