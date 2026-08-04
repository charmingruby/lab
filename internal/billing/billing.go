package billing

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/client/console"
	"github.com/charmingruby/new/internal/billing/http"
	"github.com/charmingruby/new/internal/billing/http/handler"
	"github.com/charmingruby/new/internal/billing/repository/postgres"
	"github.com/charmingruby/new/internal/billing/service"
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

	catalogService := service.NewCatalogService(offeringRepo)

	paymentService := service.NewPaymentService(txManager, offeringRepo, paymentRepo, console.NewNotifier())

	h := handler.New(catalogService, paymentService)

	http.RegisterRoutes(r, h)

	return nil
}
