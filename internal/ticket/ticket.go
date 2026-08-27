package ticket

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/lab/internal/ticket/client/console"
	"github.com/charmingruby/lab/internal/ticket/http"
	"github.com/charmingruby/lab/internal/ticket/repository/postgres"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func New(
	r chi.Router,
	db *sqlx.DB,
) error {
	txManager := postgres.NewTransactionManager(db)

	ticketRepo, err := postgres.NewTicketRepository(db)
	if err != nil {
		return err
	}

	ep := http.SetupEndpoints(
		usecase.NewCreateTicketUsecase(ticketRepo),
		usecase.NewAssignTicketUsecase(txManager, console.NewNotifier()),
		usecase.NewGetTicketUsecase(ticketRepo),
		usecase.NewListTicketsUsecase(ticketRepo),
	)

	http.RegisterRoutes(
		r, ep,
	)

	return nil
}
