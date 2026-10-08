package ticket

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/lab/internal/ticket/client/console"
	"github.com/charmingruby/lab/internal/ticket/public"
	"github.com/charmingruby/lab/internal/ticket/repository/postgres"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := postgres.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, postgres.NewTransactionManager(db), console.NewNotifier())

	return public.NewTicketReader(uc), nil
}
