package handler

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type ListPaymentsResponse struct {
	Payments []model.Payment `json:"payments"`
	Total    int             `json:"total"`
}

func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := httpx.GetQueryParam(r, "user_id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	page := httpx.GetPageQueryParam(r)

	output, err := h.paymentService.ListPayments(ctx, service.ListPaymentsInput{
		UserID: userID,
		Page:   page,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteOKResponse(w, ListPaymentsResponse{
		Payments: output.Payments,
		Total:    output.Total,
	})
}
