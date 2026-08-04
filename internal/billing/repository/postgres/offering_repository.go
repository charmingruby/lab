package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/pkg/postgrex"
)

const (
	createOfferingQuery     = "create offering"
	findOfferingByIDQuery   = "find offering by id"
	findOfferingByNameQuery = "find offering by name"
	listOfferingsQuery      = "list offerings"
	updateOfferingQuery     = "update offering"
)

var offeringQueries = map[string]string{
	createOfferingQuery: `
		INSERT INTO offerings
		(id, name, description, charge_type, price, currency, is_active, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
	findOfferingByIDQuery: `
		SELECT * FROM offerings
		WHERE
			id=$1 AND
			deleted_at IS NULL`,
	findOfferingByNameQuery: `
		SELECT * FROM offerings
		WHERE
			name=$1 AND
			deleted_at IS NULL`,
	listOfferingsQuery: `
		SELECT * FROM offerings
		WHERE
			deleted_at IS NULL
		ORDER BY created_at DESC`,
	updateOfferingQuery: `
		UPDATE offerings
		SET name=$1, description=$2, charge_type=$3, price=$4, currency=$5, is_active=$6, updated_at=$7, deleted_at=$8
		WHERE
			id=$9 AND
			deleted_at IS NULL`,
}

type OfferingRepository struct {
	db    postgrex.Querier
	stmts map[string]*sqlx.Stmt
}

func NewOfferingRepository(db postgrex.Querier) (*OfferingRepository, error) {
	stmts := make(map[string]*sqlx.Stmt, len(offeringQueries))

	for queryName, query := range offeringQueries {
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

	return &OfferingRepository{
		db:    db,
		stmts: stmts,
	}, nil
}

func (r *OfferingRepository) statement(queryName string) (*sqlx.Stmt, error) {
	stmt, ok := r.stmts[queryName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", postgrex.ErrQueryPreparation, queryName)
	}

	return stmt, nil
}

func (r *OfferingRepository) Create(ctx context.Context, offering *model.Offering) error {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(createOfferingQuery)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		offering.ID,
		offering.Name,
		offering.Description,
		offering.ChargeType,
		offering.Price,
		offering.Currency,
		offering.IsActive,
		offering.CreatedAt,
	)

	return err
}

func (r *OfferingRepository) FindByID(ctx context.Context, id string) (*model.Offering, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findOfferingByIDQuery)
	if err != nil {
		return nil, err
	}

	var offering model.Offering
	if err := stmt.QueryRowxContext(ctx, id).StructScan(&offering); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &offering, nil
}

func (r *OfferingRepository) FindByName(ctx context.Context, name string) (*model.Offering, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findOfferingByNameQuery)
	if err != nil {
		return nil, err
	}

	var offering model.Offering
	if err := stmt.QueryRowxContext(ctx, name).StructScan(&offering); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &offering, nil
}

func (r *OfferingRepository) ListAll(ctx context.Context) ([]model.Offering, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(listOfferingsQuery)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryxContext(ctx)
	if err != nil {
		return nil, err
	}

	var offerings []model.Offering
	for rows.Next() {
		var o model.Offering
		if err := rows.StructScan(&o); err != nil {
			return nil, err
		}

		offerings = append(offerings, o)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return offerings, nil
}

func (r *OfferingRepository) Update(ctx context.Context, offering *model.Offering) error {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(updateOfferingQuery)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		offering.Name,
		offering.Description,
		offering.ChargeType,
		offering.Price,
		offering.Currency,
		offering.IsActive,
		offering.UpdatedAt,
		offering.DeletedAt,
		offering.ID,
	)

	return err
}
