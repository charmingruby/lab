package endpoint

import (
	"net/http"
	"time"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type PaymentV1Response struct {
	CreatedAt     time.Time `json:"created_at"`
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	OfferingID    string    `json:"offering_id"`
	Status        string    `json:"status"`
	ExternalID    string    `json:"external_id"`
	ChargedAmount int       `json:"charged_amount"`
}

func newPaymentV1Response(payment *model.Payment) PaymentV1Response {
	return PaymentV1Response{
		ID:            payment.ID,
		UserID:        payment.UserID,
		OfferingID:    payment.OfferingID,
		Status:        payment.Status,
		ExternalID:    payment.ExternalID,
		ChargedAmount: payment.ChargedAmount,
		CreatedAt:     payment.CreatedAt,
	}
}

func (e *Endpoint) GetPaymentV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentID, err := httpx.GetPathParam(r, "id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	payment, err := e.getPayment.GetPayment(ctx, usecase.GetPaymentInput{
		PaymentID: paymentID,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteOKResponse(w, newPaymentV1Response(payment))
}
