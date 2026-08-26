package ticket

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/ticket/client/console"
	"github.com/charmingruby/new/internal/ticket/http"
	"github.com/charmingruby/new/internal/ticket/repository/postgres"
	"github.com/charmingruby/new/internal/ticket/usecase"
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
