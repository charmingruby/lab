package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type CreateOfferingV1Request struct {
	Name        string `json:"name"        validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	ChargeType  string `json:"charge_type" validate:"required,min=1"`
	Currency    string `json:"currency"    validate:"required,min=1"`
	Price       int    `json:"price"       validate:"required,gt=0"`
	IsActive    bool   `json:"is_active"`
}

type CreateOfferingV1Response struct {
	ID string `json:"id"`
}

func (e *Endpoint) CreateOfferingV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreateOfferingV1Request](w, r)
	if err != nil {
		return
	}

	output, err := e.createOffering.CreateOffering(ctx, usecase.CreateOfferingInput{
		Name:        request.Name,
		Description: request.Description,
		ChargeType:  request.ChargeType,
		Price:       request.Price,
		Currency:    request.Currency,
		IsActive:    request.IsActive,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteCreatedResponse(w, CreateOfferingV1Response{
		ID: output.ID,
	})
}
