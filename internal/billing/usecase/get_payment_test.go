package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestGetPaymentUsecase(t *testing.T) {
	tests := []struct {
		setupMocks func(*mocks.PaymentRepository)
		assert     func(*testing.T, *model.Payment, error)
		name       string
		input      usecase.GetPaymentInput
	}{
		{
			name:  "success",
			input: usecase.GetPaymentInput{PaymentID: "payment-123"},
			setupMocks: func(paymentRepo *mocks.PaymentRepository) {
				paymentRepo.On("FindByID", mock.Anything, "payment-123").Return(
					&model.Payment{Model: core.NewModel(), Status: model.PaidPaymentStatus}, nil,
				)
			},
			assert: func(t *testing.T, payment *model.Payment, err error) {
				require.NoError(t, err)
				require.NotNil(t, payment)
				assert.Equal(t, model.PaidPaymentStatus, payment.Status)
			},
		},
		{
			name:  "error when payment not found",
			input: usecase.GetPaymentInput{PaymentID: "unknown"},
			setupMocks: func(paymentRepo *mocks.PaymentRepository) {
				paymentRepo.On("FindByID", mock.Anything, "unknown").Return(nil, nil)
			},
			assert: func(t *testing.T, _ *model.Payment, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsNotFound(err))
			},
		},
		{
			name:  "error on repository failure",
			input: usecase.GetPaymentInput{PaymentID: "payment-123"},
			setupMocks: func(paymentRepo *mocks.PaymentRepository) {
				paymentRepo.On("FindByID", mock.Anything, "payment-123").Return(nil, errors.New("db error"))
			},
			assert: func(t *testing.T, _ *model.Payment, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepo := new(mocks.PaymentRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(paymentRepo)
			}

			uc := usecase.NewGetPaymentUsecase(paymentRepo)
			payment, err := uc.GetPayment(context.Background(), tt.input)

			if tt.assert != nil {
				tt.assert(t, payment, err)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
