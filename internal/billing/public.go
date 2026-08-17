package billing

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/client"
	"github.com/charmingruby/new/internal/billing/public"
	"github.com/charmingruby/new/internal/billing/repository/postgres"
	"github.com/charmingruby/new/internal/billing/usecase"
)

func NewPaymentReader(db *sqlx.DB) (client.PaymentReader, error) {
	paymentRepo, err := postgres.NewPaymentRepository(db)
	if err != nil {
		return nil, err
	}

	return public.NewPaymentReader(usecase.NewGetPaymentUsecase(paymentRepo)), nil
}
