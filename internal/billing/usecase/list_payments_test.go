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
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestListPaymentsUsecase(t *testing.T) {
	tests := []struct {
		setupMocks func(*mocks.PaymentRepository)
		assert     func(*testing.T, usecase.ListPaymentsOutput, error)
		name       string
		input      usecase.ListPaymentsInput
	}{
		{
			name: "success with default pagination",
			input: usecase.ListPaymentsInput{
				UserID: "user-123",
				Page:   0,
			},
			setupMocks: func(paymentRepo *mocks.PaymentRepository) {
				paymentRepo.On("ListByUserID", mock.Anything, "user-123", mock.Anything).
					Return([]model.Payment{
						{UserID: "user-123", Status: model.PaidPaymentStatus},
					}, 1, nil)
			},
			assert: func(t *testing.T, out usecase.ListPaymentsOutput, err error) {
				require.NoError(t, err)
				assert.Len(t, out.Payments, 1)
				assert.Equal(t, 1, out.Total)
			},
		},
		{
			name: "success with custom page",
			input: usecase.ListPaymentsInput{
				UserID: "user-123",
				Page:   2,
			},
			setupMocks: func(paymentRepo *mocks.PaymentRepository) {
				paymentRepo.On("ListByUserID", mock.Anything, "user-123", mock.Anything).
					Return([]model.Payment{}, 0, nil)
			},
			assert: func(t *testing.T, out usecase.ListPaymentsOutput, err error) {
				require.NoError(t, err)
				assert.Empty(t, out.Payments)
				assert.Equal(t, 0, out.Total)
			},
		},
		{
			name: "error on repository failure",
			input: usecase.ListPaymentsInput{
				UserID: "user-123",
				Page:   1,
			},
			setupMocks: func(paymentRepo *mocks.PaymentRepository) {
				paymentRepo.On("ListByUserID", mock.Anything, "user-123", mock.Anything).
					Return(nil, 0, errors.New("db error"))
			},
			assert: func(t *testing.T, _ usecase.ListPaymentsOutput, err error) {
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

			uc := usecase.NewListPaymentsUsecase(paymentRepo)
			output, err := uc.ListPayments(context.Background(), tt.input)

			if tt.assert != nil {
				tt.assert(t, output, err)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
