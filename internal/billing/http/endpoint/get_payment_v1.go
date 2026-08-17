package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/httpx"
)

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

	httpx.WriteOKResponse(w, payment)
}
