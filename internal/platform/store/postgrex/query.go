package postgrex

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const (
	DefaultWriteTimeout = 6 * time.Second
	DefaultReadTimeout  = 4 * time.Second
)

func One[T any](ctx context.Context, stmts *Statements, name string, args ...any) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultReadTimeout)
	defer cancel()

	stmt, err := stmts.Stmt(name)
	if err != nil {
		return nil, err
	}

	var row T
	if err := stmt.QueryRowxContext(ctx, args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &row, nil
}

func Many[T any](ctx context.Context, stmts *Statements, name string, args ...any) ([]T, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultReadTimeout)
	defer cancel()

	stmt, err := stmts.Stmt(name)
	if err != nil {
		return nil, err
	}

	rows := make([]T, 0)
	if err := stmt.SelectContext(ctx, &rows, args...); err != nil {
		return nil, err
	}

	return rows, nil
}

func Exec(ctx context.Context, stmts *Statements, name string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultWriteTimeout)
	defer cancel()

	stmt, err := stmts.Stmt(name)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, args...)

	return err
}

func ExecReturning[T any](ctx context.Context, stmts *Statements, name string, args ...any) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultWriteTimeout)
	defer cancel()

	var value T

	stmt, err := stmts.Stmt(name)
	if err != nil {
		return value, err
	}

	if err := stmt.QueryRowxContext(ctx, args...).Scan(&value); err != nil {
		return value, err
	}

	return value, nil
}
