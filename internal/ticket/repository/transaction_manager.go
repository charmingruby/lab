package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/lab/internal/platform/postgrex"
)

// Transaction groups the repositories available inside a transaction.
type Transaction struct {
	TicketRepo *TicketRepository
}

// TransactionManager runs multi-repo writes atomically. Concrete like the
// repository itself — the golden source has no port interface.
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
