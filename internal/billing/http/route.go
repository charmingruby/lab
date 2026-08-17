package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/new/internal/billing/http/endpoint"
	"github.com/charmingruby/new/internal/billing/usecase"
)

func SetupEndpoints(
	createOffering usecase.CreateOfferingUsecase,
	createPayment usecase.CreatePaymentUsecase,
	getPayment usecase.GetPaymentUsecase,
	listPayments usecase.ListPaymentsUsecase,
) *endpoint.Endpoint {
	return endpoint.New(
		createOffering,
		createPayment,
		getPayment,
		listPayments,
	)
}

func RegisterRoutes(
	r chi.Router,
	ep *endpoint.Endpoint,
) {
	r.Route("/v1/offerings", func(r chi.Router) {
		r.Post("/", ep.CreateOfferingV1)
	})

	r.Route("/v1/payments", func(r chi.Router) {
		r.Post("/", ep.CreatePaymentV1)
		r.Get("/", ep.ListPaymentsV1)
		r.Get("/{id}", ep.GetPaymentV1)
	})
}
