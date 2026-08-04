package billing

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/client"
	"github.com/charmingruby/new/internal/billing/client/console"
	"github.com/charmingruby/new/internal/billing/public"
	"github.com/charmingruby/new/internal/billing/repository/postgres"
	"github.com/charmingruby/new/internal/billing/service"
)

// NewPaymentReader assembles a read adapter for this module's payment data.
// Other modules consume it only through the client.PaymentReader port.
func NewPaymentReader(db *sqlx.DB) (client.PaymentReader, error) {
	txManager := postgres.NewTransactionManager(db)

	offeringRepo, err := postgres.NewOfferingRepository(db)
	if err != nil {
		return nil, err
	}

	paymentRepo, err := postgres.NewPaymentRepository(db)
	if err != nil {
		return nil, err
	}

	paymentService := service.NewPaymentService(txManager, offeringRepo, paymentRepo, console.NewNotifier())

	return public.NewPaymentReader(paymentService), nil
}
