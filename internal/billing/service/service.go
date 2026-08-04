package service

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (CreatePaymentOutput, error)
	GetPayment(ctx context.Context, input GetPaymentInput) (*model.Payment, error)
	ListPayments(ctx context.Context, input ListPaymentsInput) (ListPaymentsOutput, error)
}

type CatalogService interface {
	CreateOffering(ctx context.Context, input CreateOfferingInput) (CreateOfferingOutput, error)
}

type CreateOfferingInput = model.OfferingInput

type CreateOfferingOutput struct {
	ID string
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

type GetPaymentInput struct {
	PaymentID string
}

type ListPaymentsInput struct {
	UserID string
	Page   int
}

type ListPaymentsOutput struct {
	Payments []model.Payment
	Total    int
}
