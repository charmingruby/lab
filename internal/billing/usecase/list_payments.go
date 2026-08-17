package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/internal/shared/customerr"
)

type ListPaymentsUsecase interface {
	ListPayments(ctx context.Context, input ListPaymentsInput) (ListPaymentsOutput, error)
}

type ListPaymentsInput struct {
	UserID string
	Page   int
}

type ListPaymentsOutput struct {
	Payments []model.Payment
	Total    int
}

type listPaymentsUsecase struct {
	paymentRepo repository.PaymentRepository
}

func NewListPaymentsUsecase(paymentRepo repository.PaymentRepository) *listPaymentsUsecase {
	return &listPaymentsUsecase{
		paymentRepo: paymentRepo,
	}
}

func (u *listPaymentsUsecase) ListPayments(ctx context.Context, input ListPaymentsInput) (ListPaymentsOutput, error) {
	params := core.DefaultPaginationParams(input.Page)

	payments, total, err := u.paymentRepo.ListByUserID(ctx, input.UserID, params)
	if err != nil {
		return ListPaymentsOutput{}, customerr.Integration(err)
	}

	return ListPaymentsOutput{
		Payments: payments,
		Total:    total,
	}, nil
}
