package repository

import (
	"context"

	"github.com/charmingruby/lab/internal/platform/store/postgrex"
	"github.com/jmoiron/sqlx"
)

type Transaction struct {
	TicketRepo *TicketRepository
}

type TransactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (t *TransactionManager) Transact(ctx context.Context, fn func(Transaction) error) error {
	return postgrex.RunInTx(ctx, t.db, func(tx *sqlx.Tx) error {
		ticketRepo, err := NewTicketRepository(tx)
		if err != nil {
			return err
		}

		return fn(Transaction{
			TicketRepo: ticketRepo,
		})
	})
}
