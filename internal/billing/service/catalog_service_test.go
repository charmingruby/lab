package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestCatalogService_CreateOffering(t *testing.T) {
	tests := []struct {
		setupMocks func(*mocks.OfferingRepository)
		assert     func(*testing.T, service.CreateOfferingOutput, error)
		name       string
		input      service.CreateOfferingInput
	}{
		{
			name: "success",
			input: model.OfferingInput{
				Name: "Premium Plan", Description: "Premium plan",
				ChargeType: "one_time", Currency: "USD", Price: 2999, IsActive: true,
			},
			setupMocks: func(offeringRepo *mocks.OfferingRepository) {
				offeringRepo.On("FindByName", mock.Anything, "Premium Plan").Return(nil, nil)
				offeringRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Offering")).Return(nil)
			},
			assert: func(t *testing.T, out service.CreateOfferingOutput, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, out.ID)
			},
		},
		{
			name: "error when offering already exists",
			input: model.OfferingInput{
				Name: "Existing Plan", Description: "Existing",
				ChargeType: "one_time", Currency: "USD", Price: 100,
			},
			setupMocks: func(offeringRepo *mocks.OfferingRepository) {
				existing, _ := model.NewOffering(model.OfferingInput{
					Name: "Existing Plan", Description: "Existing",
					ChargeType: "one_time", Currency: "USD", Price: 100,
				})
				offeringRepo.On("FindByName", mock.Anything, "Existing Plan").Return(existing, nil)
			},
			assert: func(t *testing.T, _ service.CreateOfferingOutput, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsConflict(err))
			},
		},
		{
			name: "error on FindByName integration failure",
			input: model.OfferingInput{
				Name: "Test", Description: "Test",
				ChargeType: "one_time", Currency: "USD", Price: 100,
			},
			setupMocks: func(offeringRepo *mocks.OfferingRepository) {
				offeringRepo.On("FindByName", mock.Anything, "Test").Return(nil, errors.New("db connection failed"))
			},
			assert: func(t *testing.T, _ service.CreateOfferingOutput, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsInvalidOperation(err))
			},
		},
		{
			name: "error on validation failure",
			input: model.OfferingInput{
				Name: "Test", Description: "Test",
				ChargeType: "invalid", Currency: "USD", Price: 100,
			},
			setupMocks: func(offeringRepo *mocks.OfferingRepository) {
				offeringRepo.On("FindByName", mock.Anything, "Test").Return(nil, nil)
			},
			assert: func(t *testing.T, _ service.CreateOfferingOutput, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsValidation(err))
			},
		},
		{
			name: "error on Create integration failure",
			input: model.OfferingInput{
				Name: "Test", Description: "Test",
				ChargeType: "one_time", Currency: "USD", Price: 100,
			},
			setupMocks: func(offeringRepo *mocks.OfferingRepository) {
				offeringRepo.On("FindByName", mock.Anything, "Test").Return(nil, nil)
				offeringRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Offering")).
					Return(errors.New("insert failed"))
			},
			assert: func(t *testing.T, _ service.CreateOfferingOutput, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsInvalidOperation(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offeringRepo := new(mocks.OfferingRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(offeringRepo)
			}

			svc := service.NewCatalogService(offeringRepo)
			output, err := svc.CreateOffering(context.Background(), tt.input)

			if tt.assert != nil {
				tt.assert(t, output, err)
			}

			offeringRepo.AssertExpectations(t)
		})
	}
}
