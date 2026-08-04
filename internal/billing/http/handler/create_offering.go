package handler

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type CreateOfferingRequest struct {
	Name        string `json:"name"        validate:"required,min=1"`
	Description string `json:"description" validate:"required,min=1"`
	ChargeType  string `json:"charge_type" validate:"required,min=1"`
	Currency    string `json:"currency"    validate:"required,min=1"`
	Price       int    `json:"price"       validate:"required"`
	IsActive    bool   `json:"is_active"`
}

type CreateOfferingResponse struct {
	ID string `json:"id"`
}

func (h *Handler) CreateOffering(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := httpx.ParseRequest[CreateOfferingRequest](w, r)
	if err != nil {
		return
	}

	output, err := h.catalogService.CreateOffering(ctx, service.CreateOfferingInput{
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

	httpx.WriteCreatedResponse(w, CreateOfferingResponse{
		ID: output.ID,
	})
}
