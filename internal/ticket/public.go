package ticket

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/lab/internal/ticket/client/console"
	"github.com/charmingruby/lab/internal/ticket/public"
	"github.com/charmingruby/lab/internal/ticket/repository"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func NewTicketReader(db *sqlx.DB) (public.Reader, error) {
	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		return nil, err
	}

	uc := usecase.New(ticketRepo, repository.NewTransactionManager(db), console.NewNotifier())

	return public.NewTicketReader(uc), nil
}
