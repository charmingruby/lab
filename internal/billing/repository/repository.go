package repository

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/shared/core"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	FindByID(ctx context.Context, id string) (*model.Payment, error)
	FindByExternalID(ctx context.Context, externalID string) (*model.Payment, error)
	ListByUserID(ctx context.Context, userID string, params core.PaginationParams) ([]model.Payment, int, error)
	Update(ctx context.Context, payment *model.Payment) error
}

type OfferingRepository interface {
	Create(ctx context.Context, offering *model.Offering) error
	FindByID(ctx context.Context, id string) (*model.Offering, error)
	FindByName(ctx context.Context, name string) (*model.Offering, error)
	ListAll(ctx context.Context) ([]model.Offering, error)
	Update(ctx context.Context, offering *model.Offering) error
}

type Transaction struct {
	PaymentRepo  PaymentRepository
	OfferingRepo OfferingRepository
}
