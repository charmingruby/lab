package ticket

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/lab/internal/ticket/client/notifier"
	"github.com/charmingruby/lab/internal/ticket/http"
	"github.com/charmingruby/lab/internal/ticket/public"
	"github.com/charmingruby/lab/internal/ticket/repository"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func New(
	r chi.Router,
	db *sqlx.DB,
) error {
	txManager := repository.NewTransactionManager(db)

	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return err
	}

	uc := usecase.New(ticketRepo, txManager, notifier.NewConsole())

	ep := http.SetupEndpoints(uc)

	http.RegisterRoutes(
		r, ep,
	)

	return nil
}

func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, repository.NewTransactionManager(db), notifier.NewConsole())

	return public.NewTicketReader(uc), nil
}
