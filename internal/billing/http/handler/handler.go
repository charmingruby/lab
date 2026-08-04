package handler

import "github.com/charmingruby/new/internal/billing/service"

type Handler struct {
	catalogService service.CatalogService
	paymentService service.PaymentService
}

func New(
	catalogService service.CatalogService,
	paymentService service.PaymentService,
) *Handler {
	return &Handler{
		catalogService: catalogService,
		paymentService: paymentService,
	}
}
