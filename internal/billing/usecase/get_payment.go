package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/internal/shared/customerr"
)

type GetPaymentInput struct {
	PaymentID string
}

type getPaymentUsecase struct {
	paymentRepo repository.PaymentRepository
}

func NewGetPaymentUsecase(paymentRepo repository.PaymentRepository) *getPaymentUsecase {
	return &getPaymentUsecase{
		paymentRepo: paymentRepo,
	}
}

func (u *getPaymentUsecase) GetPayment(ctx context.Context, input GetPaymentInput) (*model.Payment, error) {
	payment, err := u.paymentRepo.FindByID(ctx, input.PaymentID)
	if err != nil {
		return nil, customerr.Integration(err)
	}

	if payment == nil {
		return nil, customerr.NotFound("payment not found")
	}

	return payment, nil
}
