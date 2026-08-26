package usecase

import (
	"context"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/internal/ticket/repository"
	"github.com/charmingruby/lab/pkg/o11y"
)

type AssignTicketInput struct {
	TicketID   string
	AssigneeID string
}

type assignTicketUsecase struct {
	txManager core.TransactionManager[repository.Transaction]
	notifier  client.NotificationClient
}

func NewAssignTicketUsecase(
	txManager core.TransactionManager[repository.Transaction],
	notifier client.NotificationClient,
) *assignTicketUsecase {
	return &assignTicketUsecase{
		txManager: txManager,
		notifier:  notifier,
	}
}

func (u *assignTicketUsecase) AssignTicket(
	ctx context.Context,
	input AssignTicketInput,
) error {
	var assigneeID string

	err := u.txManager.Transact(func(tx repository.Transaction) error {
		ticket, err := tx.TicketRepo.FindByID(ctx, input.TicketID)
		if err != nil {
			return customerr.Integration(err)
		}

		if ticket == nil {
			return customerr.NotFound("ticket not found")
		}

		if err := ticket.Assign(input.AssigneeID); err != nil {
			return customerr.Validation(err.Error())
		}

		if err := tx.TicketRepo.Update(ctx, ticket); err != nil {
			return customerr.Integration(err)
		}

		assigneeID = input.AssigneeID

		return nil
	})
	if err != nil {
		return err
	}

	if err := u.notifier.Send(ctx, client.SendNotificationInput{
		AssigneeID: assigneeID,
		Message:    "you have been assigned to a ticket",
	}); err != nil {
		o11y.LoggerFromContext(ctx).Warn("notification send failed", "error", err)
	}

	return nil
}
