package ticket

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/ticket/public"
	"github.com/charmingruby/new/internal/ticket/repository/postgres"
	"github.com/charmingruby/new/internal/ticket/usecase"
)

func NewTicketReader(db *sqlx.DB) (*public.TicketReader, error) {
	ticketRepo, err := postgres.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	getTicketUc := usecase.NewGetTicketUsecase(ticketRepo)

	return public.NewTicketReader(getTicketUc), nil
}
