package handler

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type CreatePaymentRequest struct {
	UserID        string `json:"user_id"        validate:"required,min=1"`
	OfferingID    string `json:"offering_id"    validate:"required,min=1"`
	ExternalID    string `json:"external_id"    validate:"required,min=1"`
	ChargedAmount int    `json:"charged_amount" validate:"required"`
}

type CreatePaymentResponse struct {
	ID string `json:"id"`
}

func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreatePaymentRequest](w, r)
	if err != nil {
		return
	}

	output, err := h.paymentService.CreatePayment(ctx, service.CreatePaymentInput{
		UserID:        request.UserID,
		OfferingID:    request.OfferingID,
		ExternalID:    request.ExternalID,
		ChargedAmount: request.ChargedAmount,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreatePaymentResponse{
		ID: output.ID,
	})
}
