package handler

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/internal/shared/httpx"
)

func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentID, err := httpx.GetPathParam(r, "id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	payment, err := h.paymentService.GetPayment(ctx, service.GetPaymentInput{
		PaymentID: paymentID,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteOKResponse(w, payment)
}
