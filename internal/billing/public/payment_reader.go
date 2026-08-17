package public

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/usecase"
)

type PaymentReader struct {
	getPayment usecase.GetPaymentUsecase
}

func NewPaymentReader(getPayment usecase.GetPaymentUsecase) *PaymentReader {
	return &PaymentReader{getPayment: getPayment}
}

func (r *PaymentReader) GetPayment(ctx context.Context, paymentID string) (*model.Payment, error) {
	return r.getPayment.GetPayment(ctx, usecase.GetPaymentInput{
		PaymentID: paymentID,
	})
}
