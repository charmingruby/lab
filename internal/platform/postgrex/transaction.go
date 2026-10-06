package postgrex

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	sqlx.ExtContext
	Preparex(query string) (*sqlx.Stmt, error)
}

func RunInTx(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		if commitErr := tx.Commit(); commitErr != nil {
			return commitErr
		}

		return nil
	}

	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
