package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/pkg/postgrex"
)

const (
	createPaymentQuery           = "create payment"
	findPaymentByIDQuery         = "find payment by id"
	findPaymentByExternalIDQuery = "find payment by external id"
	listPaymentByUserIDQuery     = "list payment by user id"
	countPaymentByUserIDQuery    = "count payment by user id"
	updatePaymentQuery           = "update payment"
)

var paymentQueries = map[string]string{
	createPaymentQuery: `
		INSERT INTO payments
		(id, user_id, offering_id, status, charged_amount, external_id, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7)`,
	findPaymentByIDQuery: `
		SELECT * FROM payments
		WHERE
			id=$1 AND
			deleted_at IS NULL`,
	findPaymentByExternalIDQuery: `
		SELECT * FROM payments
		WHERE
			external_id=$1 AND
			deleted_at IS NULL`,
	listPaymentByUserIDQuery: `
		SELECT * FROM payments
		WHERE
			user_id=$1 AND
			deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
	countPaymentByUserIDQuery: `
		SELECT COUNT(*) FROM payments
		WHERE
			user_id=$1 AND
			deleted_at IS NULL`,
	updatePaymentQuery: `
		UPDATE payments
		SET status=$1, updated_at=$2, deleted_at=$3
		WHERE
			id=$4 AND
			deleted_at IS NULL`,
}

type PaymentRepository struct {
	db    postgrex.Querier
	stmts map[string]*sqlx.Stmt
}

func NewPaymentRepository(db postgrex.Querier) (*PaymentRepository, error) {
	stmts := make(map[string]*sqlx.Stmt, len(paymentQueries))

	for queryName, query := range paymentQueries {
		stmt, err := db.Preparex(query)
		if err != nil {
			return nil, fmt.Errorf("%w (%s): %s",
				postgrex.ErrQueryPreparation,
				queryName,
				err.Error(),
			)
		}

		stmts[queryName] = stmt
	}

	return &PaymentRepository{
		db:    db,
		stmts: stmts,
	}, nil
}

func (r *PaymentRepository) statement(queryName string) (*sqlx.Stmt, error) {
	stmt, ok := r.stmts[queryName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", postgrex.ErrQueryPreparation, queryName)
	}

	return stmt, nil
}

func (r *PaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(createPaymentQuery)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		payment.ID,
		payment.UserID,
		payment.OfferingID,
		payment.Status,
		payment.ChargedAmount,
		payment.ExternalID,
		payment.CreatedAt,
	)

	return err
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*model.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findPaymentByIDQuery)
	if err != nil {
		return nil, err
	}

	var payment model.Payment
	if err := stmt.QueryRowxContext(ctx, id).StructScan(&payment); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) FindByExternalID(ctx context.Context, externalID string) (*model.Payment, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findPaymentByExternalIDQuery)
	if err != nil {
		return nil, err
	}

	var payment model.Payment
	if err := stmt.QueryRowxContext(ctx, externalID).StructScan(&payment); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) ListByUserID(
	ctx context.Context,
	userID string,
	params core.PaginationParams,
) ([]model.Payment, int, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(listPaymentByUserIDQuery)
	if err != nil {
		return nil, 0, err
	}

	rows, err := stmt.QueryxContext(ctx, userID, params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		return nil, 0, err
	}

	var payments []model.Payment
	for rows.Next() {
		var p model.Payment
		if err := rows.StructScan(&p); err != nil {
			return nil, 0, err
		}

		payments = append(payments, p)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	countStmt, err := r.statement(countPaymentByUserIDQuery)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := countStmt.QueryRowxContext(ctx, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(updatePaymentQuery)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		payment.Status,
		payment.UpdatedAt,
		payment.DeletedAt,
		payment.ID,
	)

	return err
}
