package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/internal/ticket/model"
	"github.com/charmingruby/new/pkg/postgrex"
)

const (
	createTicketQuery        = "create ticket"
	findTicketByIDQuery      = "find ticket by id"
	updateTicketQuery        = "update ticket"
	listTicketByStatusQuery  = "list ticket by status"
	countTicketByStatusQuery = "count ticket by status"
)

var ticketQueries = map[string]string{
	createTicketQuery: `
		INSERT INTO tickets
		(id, title, description, status, priority, created_at)
		VALUES($1, $2, $3, $4, $5, $6)`,
	findTicketByIDQuery: `
		SELECT * FROM tickets
		WHERE
			id=$1 AND
			deleted_at IS NULL`,
	updateTicketQuery: `
		UPDATE tickets
		SET status=$2, assignee_id=$3, updated_at=$4
		WHERE
			id=$1 AND
			deleted_at IS NULL`,
	listTicketByStatusQuery: `
		SELECT * FROM tickets
		WHERE
			status=$1 AND
			deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
	countTicketByStatusQuery: `
		SELECT COUNT(*) FROM tickets
		WHERE
			status=$1 AND
			deleted_at IS NULL`,
}

type TicketRepository struct {
	db    postgrex.Querier
	stmts map[string]*sqlx.Stmt
}

func NewTicketRepository(db postgrex.Querier) (*TicketRepository, error) {
	stmts := make(map[string]*sqlx.Stmt, len(ticketQueries))

	for queryName, query := range ticketQueries {
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

	return &TicketRepository{
		db:    db,
		stmts: stmts,
	}, nil
}

func (r *TicketRepository) statement(queryName string) (*sqlx.Stmt, error) {
	stmt, ok := r.stmts[queryName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", postgrex.ErrQueryPreparation, queryName)
	}

	return stmt, nil
}

func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(createTicketQuery)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		ticket.ID,
		ticket.Title,
		ticket.Description,
		ticket.Status,
		ticket.Priority,
		ticket.CreatedAt,
	)

	return err
}

func (r *TicketRepository) FindByID(ctx context.Context, id string) (*model.Ticket, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(findTicketByIDQuery)
	if err != nil {
		return nil, err
	}

	var ticket model.Ticket
	if err := stmt.QueryRowxContext(ctx, id).StructScan(&ticket); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(updateTicketQuery)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx,
		ticket.ID,
		ticket.Status,
		ticket.AssigneeID,
		ticket.UpdatedAt,
	)

	return err
}

func (r *TicketRepository) ListByStatus(
	ctx context.Context,
	status string,
	params core.PaginationParams,
) ([]model.Ticket, int, error) {
	ctx, cancel := context.WithTimeout(ctx, postgrex.DefaultReadTimeout)
	defer cancel()

	stmt, err := r.statement(listTicketByStatusQuery)
	if err != nil {
		return nil, 0, err
	}

	rows, err := stmt.QueryxContext(ctx, status, params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		return nil, 0, err
	}

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.StructScan(&t); err != nil {
			return nil, 0, err
		}

		tickets = append(tickets, t)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	countStmt, err := r.statement(countTicketByStatusQuery)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := countStmt.QueryRowxContext(ctx, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}
