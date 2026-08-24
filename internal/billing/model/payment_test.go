package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
)

func TestNewPayment(t *testing.T) {
	validInput := model.PaymentInput{
		UserID:        "user-123",
		OfferingID:    "offering-123",
		ExternalID:    "pi_mock_123",
		ChargedAmount: 2999,
	}

	tests := []struct {
		arrange func(t *testing.T) *model.Payment
		assert  func(t *testing.T, p *model.Payment)
		name    string
	}{
		{
			name: "should create successfully with all fields",
			arrange: func(t *testing.T) *model.Payment {
				return model.NewPayment(validInput)
			},
			assert: func(t *testing.T, p *model.Payment) {
				require.NotNil(t, p)

				assert.NotEmpty(t, p.ID)
				assert.Equal(t, validInput.UserID, p.UserID)
				assert.Equal(t, validInput.OfferingID, p.OfferingID)
				assert.Equal(t, validInput.ExternalID, p.ExternalID)
				assert.Equal(t, validInput.ChargedAmount, p.ChargedAmount)
				assert.Equal(t, model.PendingPaymentStatus, p.Status)

				assert.NotZero(t, p.CreatedAt)
				assert.Nil(t, p.UpdatedAt)
				assert.Nil(t, p.DeletedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := tt.arrange(t)
			tt.assert(t, payment)
		})
	}
}

func TestPayment_MarkAsPaid(t *testing.T) {
	payment := model.NewPayment(model.PaymentInput{
		UserID: "user-123", OfferingID: "offering-123",
		ExternalID: "pi_123", ChargedAmount: 1000,
	})
	require.Equal(t, model.PendingPaymentStatus, payment.Status)

	require.NoError(t, payment.MarkAsPaid())

	assert.Equal(t, model.PaidPaymentStatus, payment.Status)
	assert.NotNil(t, payment.UpdatedAt)
	assert.Nil(t, payment.DeletedAt)
}

func TestPayment_MarkAsFailed(t *testing.T) {
	payment := model.NewPayment(model.PaymentInput{
		UserID: "user-123", OfferingID: "offering-123",
		ExternalID: "pi_123", ChargedAmount: 1000,
	})
	require.Equal(t, model.PendingPaymentStatus, payment.Status)

	require.NoError(t, payment.MarkAsFailed())

	assert.Equal(t, model.FailedPaymentStatus, payment.Status)
	assert.NotNil(t, payment.UpdatedAt)
	assert.Nil(t, payment.DeletedAt)
}

func TestPayment_InvalidTransitions(t *testing.T) {
	tests := []struct {
		arrange func(t *testing.T) *model.Payment
		act     func(*model.Payment) error
		name    string
	}{
		{
			name: "paid -> paid",
			arrange: func(t *testing.T) *model.Payment {
				payment := model.NewPayment(model.PaymentInput{
					UserID: "user-123", OfferingID: "offering-123",
					ExternalID: "pi_123", ChargedAmount: 1000,
				})
				require.NoError(t, payment.MarkAsPaid())

				return payment
			},
			act: (*model.Payment).MarkAsPaid,
		},
		{
			name: "paid -> failed",
			arrange: func(t *testing.T) *model.Payment {
				payment := model.NewPayment(model.PaymentInput{
					UserID: "user-123", OfferingID: "offering-123",
					ExternalID: "pi_123", ChargedAmount: 1000,
				})
				require.NoError(t, payment.MarkAsPaid())

				return payment
			},
			act: (*model.Payment).MarkAsFailed,
		},
		{
			name: "failed -> paid",
			arrange: func(t *testing.T) *model.Payment {
				payment := model.NewPayment(model.PaymentInput{
					UserID: "user-123", OfferingID: "offering-123",
					ExternalID: "pi_123", ChargedAmount: 1000,
				})
				require.NoError(t, payment.MarkAsFailed())

				return payment
			},
			act: (*model.Payment).MarkAsPaid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := tt.arrange(t)

			err := tt.act(payment)

			require.ErrorIs(t, err, model.ErrInvalidPaymentTransition)
		})
	}
}
