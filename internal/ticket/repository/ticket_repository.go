package repository

import (
	"context"

	"github.com/charmingruby/lab/internal/platform/store/postgrex"
	"github.com/charmingruby/lab/internal/ticket/model"
)

const (
	createTicketQuery   = "create ticket"
	findTicketByIDQuery = "find ticket by id"
	updateTicketQuery   = "update ticket"
)

var ticketQueries = map[string]string{
	createTicketQuery: `
		INSERT INTO tickets
			(id, title, description, status, priority, created_at)
		VALUES
			($1, $2, $3, $4, $5, $6)`,
	findTicketByIDQuery: `
		SELECT * FROM tickets
		WHERE
			id = $1 AND
			deleted_at IS NULL`,
	updateTicketQuery: `
		UPDATE tickets
		SET
			status = $2,
			assignee_id = $3,
			updated_at = $4
		WHERE
			id = $1 AND
			deleted_at IS NULL`,
}

type TicketRepository struct {
	stmts *postgrex.Statements
}

func NewTicketRepository(db postgrex.Querier) (*TicketRepository, error) {
	stmts, err := postgrex.Prepare(db, ticketQueries)
	if err != nil {
		return nil, err
	}

	return &TicketRepository{stmts: stmts}, nil
}

func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	return postgrex.Exec(ctx, r.stmts, createTicketQuery,
		ticket.ID,
		ticket.Title,
		ticket.Description,
		ticket.Status,
		ticket.Priority,
		ticket.CreatedAt,
	)
}

func (r *TicketRepository) FindByID(ctx context.Context, id string) (*model.Ticket, error) {
	return postgrex.One[model.Ticket](ctx, r.stmts, findTicketByIDQuery, id)
}

func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	return postgrex.Exec(ctx, r.stmts, updateTicketQuery,
		ticket.ID,
		ticket.Status,
		ticket.AssigneeID,
		ticket.UpdatedAt,
	)
}
