package repository

import (
	"context"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/ticket/model"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	FindByID(ctx context.Context, id string) (*model.Ticket, error)
	Update(ctx context.Context, ticket *model.Ticket) error
	ListByStatus(ctx context.Context, status string, params core.PaginationParams) ([]model.Ticket, int, error)
}

type Transaction struct {
	TicketRepo TicketRepository
}
