package service

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/internal/shared/customerr"
)

type catalogService struct {
	offeringRepo repository.OfferingRepository
}

func NewCatalogService(offeringRepo repository.OfferingRepository) *catalogService {
	return &catalogService{
		offeringRepo: offeringRepo,
	}
}

func (c *catalogService) CreateOffering(ctx context.Context, input CreateOfferingInput) (CreateOfferingOutput, error) {
	offering, err := c.offeringRepo.FindByName(ctx, input.Name)
	if err != nil {
		return CreateOfferingOutput{}, customerr.Integration(err)
	}

	if offering != nil {
		return CreateOfferingOutput{}, customerr.Conflict("offering already exists")
	}

	newOffering, err := model.NewOffering(input)
	if err != nil {
		return CreateOfferingOutput{}, customerr.Validation(err.Error())
	}

	if err := c.offeringRepo.Create(ctx, newOffering); err != nil {
		return CreateOfferingOutput{}, customerr.Integration(err)
	}

	return CreateOfferingOutput{
		ID: newOffering.ID,
	}, nil
}
