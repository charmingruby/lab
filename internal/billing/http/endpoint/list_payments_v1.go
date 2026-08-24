package endpoint

import (
	"net/http"

	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/httpx"
)

type ListPaymentsV1Response struct {
	Payments []PaymentV1Response `json:"payments"`
	Total    int                 `json:"total"`
}

func (e *Endpoint) ListPaymentsV1(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := httpx.GetQueryParam(r, "user_id")
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	page := httpx.GetPageQueryParam(r)

	output, err := e.listPayments.ListPayments(ctx, usecase.ListPaymentsInput{
		UserID: userID,
		Page:   page,
	})
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	payments := make([]PaymentV1Response, 0, len(output.Payments))
	for _, p := range output.Payments {
		payments = append(payments, newPaymentV1Response(&p))
	}

	httpx.WriteOKResponse(w, ListPaymentsV1Response{
		Payments: payments,
		Total:    output.Total,
	})
}
