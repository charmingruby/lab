package public_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/public"
	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestPaymentReader_GetPayment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := new(mocks.PaymentService)
		svc.On("GetPayment", mock.Anything, service.GetPaymentInput{PaymentID: "payment-123"}).
			Return(&model.Payment{Status: model.PaidPaymentStatus}, nil)

		reader := public.NewPaymentReader(svc)

		payment, err := reader.GetPayment(context.Background(), "payment-123")

		require.NoError(t, err)
		require.NotNil(t, payment)
		assert.Equal(t, model.PaidPaymentStatus, payment.Status)
		svc.AssertExpectations(t)
	})

	t.Run("propagates not found", func(t *testing.T) {
		svc := new(mocks.PaymentService)
		svc.On("GetPayment", mock.Anything, service.GetPaymentInput{PaymentID: "unknown"}).
			Return(nil, customerr.NotFound("payment not found"))

		reader := public.NewPaymentReader(svc)

		_, err := reader.GetPayment(context.Background(), "unknown")

		require.Error(t, err)
		assert.True(t, customerr.IsNotFound(err))
	})
}
