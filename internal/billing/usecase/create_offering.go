package usecase

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/internal/shared/customerr"
)

type CreateOfferingUsecase interface {
	CreateOffering(ctx context.Context, input CreateOfferingInput) (CreateOfferingOutput, error)
}

type CreateOfferingInput = model.OfferingInput

type CreateOfferingOutput struct {
	ID string
}

type createOfferingUsecase struct {
	offeringRepo repository.OfferingRepository
}

func NewCreateOfferingUsecase(offeringRepo repository.OfferingRepository) *createOfferingUsecase {
	return &createOfferingUsecase{
		offeringRepo: offeringRepo,
	}
}

func (c *createOfferingUsecase) CreateOffering(
	ctx context.Context,
	input CreateOfferingInput,
) (CreateOfferingOutput, error) {
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
