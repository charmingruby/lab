package postgrex

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Statements struct {
	stmts map[string]*sqlx.Stmt
}

func Prepare(db Querier, queries map[string]string) (*Statements, error) {
	stmts := make(map[string]*sqlx.Stmt, len(queries))

	for name, query := range queries {
		stmt, err := db.Preparex(query)
		if err != nil {
			return nil, fmt.Errorf("%w (%s): %s", ErrQueryPreparation, name, err.Error())
		}

		stmts[name] = stmt
	}

	return &Statements{stmts: stmts}, nil
}

func (s *Statements) Stmt(name string) (*sqlx.Stmt, error) {
	stmt, ok := s.stmts[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrQueryPreparation, name)
	}

	return stmt, nil
}
