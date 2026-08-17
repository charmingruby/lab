package billing

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/client/console"
	"github.com/charmingruby/new/internal/billing/http"
	"github.com/charmingruby/new/internal/billing/repository/postgres"
	"github.com/charmingruby/new/internal/billing/usecase"
)

func New(
	r chi.Router,
	db *sqlx.DB,
) error {
	txManager := postgres.NewTransactionManager(db)

	offeringRepo, err := postgres.NewOfferingRepository(db)
	if err != nil {
		return err
	}

	paymentRepo, err := postgres.NewPaymentRepository(db)
	if err != nil {
		return err
	}

	ep := http.SetupEndpoints(
		usecase.NewCreateOfferingUsecase(offeringRepo),
		usecase.NewCreatePaymentUsecase(txManager, offeringRepo, paymentRepo, console.NewNotifier()),
		usecase.NewGetPaymentUsecase(paymentRepo),
		usecase.NewListPaymentsUsecase(paymentRepo),
	)

	http.RegisterRoutes(
		r, ep,
	)

	return nil
}
