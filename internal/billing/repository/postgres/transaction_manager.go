package postgres

import (
	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/pkg/postgrex"
)

type TransactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (t *TransactionManager) Transact(fn func(repository.Transaction) error) error {
	return postgrex.RunInTx(t.db, func(tx *sqlx.Tx) error {
		offeringRepo, err := NewOfferingRepository(tx)
		if err != nil {
			return err
		}

		paymentRepo, err := NewPaymentRepository(tx)
		if err != nil {
			return err
		}

		return fn(repository.Transaction{
			OfferingRepo: offeringRepo,
			PaymentRepo:  paymentRepo,
		})
	})
}
