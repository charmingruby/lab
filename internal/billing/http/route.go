package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/new/internal/billing/http/handler"
)

func RegisterRoutes(r chi.Router, h *handler.Handler) {
	r.Route("/v1/offerings", func(r chi.Router) {
		r.Post("/", h.CreateOffering)
	})

	r.Route("/v1/payments", func(r chi.Router) {
		r.Post("/", h.CreatePayment)
		r.Get("/", h.ListPayments)
		r.Get("/{id}", h.GetPayment)
	})
}
