package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
)

type CreateOfferingUsecase interface {
	CreateOffering(ctx context.Context, input CreateOfferingInput) (CreateOfferingOutput, error)
}

type CreatePaymentUsecase interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentOutput, error)
}

type GetPaymentUsecase interface {
	GetPayment(ctx context.Context, input GetPaymentInput) (*model.Payment, error)
}

type ListPaymentsUsecase interface {
	ListPayments(ctx context.Context, input ListPaymentsInput) (ListPaymentsOutput, error)
}
