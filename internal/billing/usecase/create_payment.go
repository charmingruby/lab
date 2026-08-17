package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/billing/client"
	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/pkg/o11y"
)

type CreatePaymentUsecase interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentOutput, error)
}

type CreatePaymentInput struct {
	UserID        string
	OfferingID    string
	ExternalID    string
	ChargedAmount int
}

type CreatePaymentOutput struct {
	ID string
}

type createPaymentUsecase struct {
	txManager    core.TransactionManager[repository.Transaction]
	offeringRepo repository.OfferingRepository
	paymentRepo  repository.PaymentRepository
	notifier     client.NotificationClient
}

func NewCreatePaymentUsecase(
	txManager core.TransactionManager[repository.Transaction],
	offeringRepo repository.OfferingRepository,
	paymentRepo repository.PaymentRepository,
	notifier client.NotificationClient,
) *createPaymentUsecase {
	return &createPaymentUsecase{
		txManager:    txManager,
		offeringRepo: offeringRepo,
		paymentRepo:  paymentRepo,
		notifier:     notifier,
	}
}

func (u *createPaymentUsecase) CreatePayment(
	ctx context.Context,
	input CreatePaymentInput,
) (CreatePaymentOutput, error) {
	offering, err := u.offeringRepo.FindByID(ctx, input.OfferingID)
	if err != nil {
		return CreatePaymentOutput{}, customerr.Integration(err)
	}

	if offering == nil {
		return CreatePaymentOutput{}, customerr.NotFound("offering not found")
	}

	var paymentID string
	created := false

	err = u.txManager.Transact(func(tx repository.Transaction) error {
		existing, err := tx.PaymentRepo.FindByExternalID(ctx, input.ExternalID)
		if err != nil {
			return customerr.Integration(err)
		}

		if existing != nil {
			paymentID = existing.ID
			return nil
		}

		payment := model.NewPayment(model.PaymentInput{
			UserID:        input.UserID,
			OfferingID:    input.OfferingID,
			ExternalID:    input.ExternalID,
			ChargedAmount: input.ChargedAmount,
		})

		payment.MarkAsPaid()

		if err := tx.PaymentRepo.Create(ctx, payment); err != nil {
			return customerr.Integration(err)
		}

		paymentID = payment.ID
		created = true

		return nil
	})
	if err != nil {
		return CreatePaymentOutput{}, err
	}

	if created {
		if err := u.notifier.Send(ctx, client.SendNotificationInput{
			UserID:  input.UserID,
			Message: "payment confirmed",
		}); err != nil {
			o11y.LoggerFromContext(ctx).Warn("notification send failed", "error", err)
		}
	}

	return CreatePaymentOutput{ID: paymentID}, nil
}
