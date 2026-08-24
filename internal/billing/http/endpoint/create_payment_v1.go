package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type CreatePaymentV1Request struct {
	UserID        string `json:"user_id"        validate:"required,min=1"`
	OfferingID    string `json:"offering_id"    validate:"required,min=1"`
	ExternalID    string `json:"external_id"    validate:"required,min=1"`
	ChargedAmount int    `json:"charged_amount" validate:"required,gt=0"`
}

type CreatePaymentV1Response struct {
	ID string `json:"id"`
}

func (e *Endpoint) CreatePaymentV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreatePaymentV1Request](w, r)
	if err != nil {
		return
	}

	output, err := e.createPayment.CreatePayment(ctx, usecase.CreatePaymentInput{
		UserID:        request.UserID,
		OfferingID:    request.OfferingID,
		ExternalID:    request.ExternalID,
		ChargedAmount: request.ChargedAmount,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreatePaymentV1Response{
		ID: output.ID,
	})
}
