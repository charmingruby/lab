package public_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/public"
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestPaymentReader_GetPayment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		getPayment := new(mocks.GetPaymentUsecase)
		getPayment.On("GetPayment", mock.Anything, usecase.GetPaymentInput{PaymentID: "payment-123"}).
			Return(&model.Payment{Status: model.PaidPaymentStatus}, nil)

		reader := public.NewPaymentReader(getPayment)

		payment, err := reader.GetPayment(context.Background(), "payment-123")

		require.NoError(t, err)
		require.NotNil(t, payment)
		assert.Equal(t, model.PaidPaymentStatus, payment.Status)
		getPayment.AssertExpectations(t)
	})

	t.Run("propagates not found", func(t *testing.T) {
		getPayment := new(mocks.GetPaymentUsecase)
		getPayment.On("GetPayment", mock.Anything, usecase.GetPaymentInput{PaymentID: "unknown"}).
			Return(nil, customerr.NotFound("payment not found"))

		reader := public.NewPaymentReader(getPayment)

		_, err := reader.GetPayment(context.Background(), "unknown")

		require.Error(t, err)
		assert.True(t, customerr.IsNotFound(err))
	})
}
