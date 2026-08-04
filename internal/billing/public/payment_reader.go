package public

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/service"
)

type PaymentReader struct {
	paymentService service.PaymentService
}

func NewPaymentReader(paymentService service.PaymentService) *PaymentReader {
	return &PaymentReader{paymentService: paymentService}
}

func (r *PaymentReader) GetPayment(ctx context.Context, paymentID string) (*model.Payment, error) {
	return r.paymentService.GetPayment(ctx, service.GetPaymentInput{
		PaymentID: paymentID,
	})
}
