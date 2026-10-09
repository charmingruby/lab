package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/lab/internal/platform/postgrex"
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

func (t *TransactionManager) Transact(fn func(Transaction) error) error {
	return postgrex.RunInTx(t.db, func(tx *sqlx.Tx) error {
		ticketRepo, err := NewTicketRepository(tx)
		if err != nil {
			return err
		}

		return fn(Transaction{
			TicketRepo: ticketRepo,
		})
	})
}
