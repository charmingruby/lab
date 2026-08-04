package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
)

func TestNewOffering(t *testing.T) {
	validInput := model.OfferingInput{
		Name:        "Premium Plan",
		Description: "Premium subscription plan",
		ChargeType:  "one_time",
		Currency:    "USD",
		Price:       2999,
		IsActive:    true,
	}

	tests := []struct {
		arrange func(t *testing.T) (*model.Offering, error)
		assert  func(t *testing.T, offering *model.Offering, err error)
		name    string
	}{
		{
			name: "should create successfully with all fields",
			arrange: func(t *testing.T) (*model.Offering, error) {
				return model.NewOffering(validInput)
			},
			assert: func(t *testing.T, o *model.Offering, err error) {
				require.NoError(t, err)
				require.NotNil(t, o)

				assert.NotEmpty(t, o.ID)
				assert.Equal(t, validInput.Name, o.Name)
				assert.Equal(t, validInput.Description, o.Description)
				assert.Equal(t, model.ChargeType(validInput.ChargeType), o.ChargeType)
				assert.Equal(t, validInput.Currency, o.Currency)
				assert.Equal(t, validInput.Price, o.Price)
				assert.Equal(t, validInput.IsActive, o.IsActive)

				assert.NotZero(t, o.CreatedAt)
				assert.Nil(t, o.UpdatedAt)
				assert.Nil(t, o.DeletedAt)
			},
		},
		{
			name: "should fail when charge type is invalid",
			arrange: func(t *testing.T) (*model.Offering, error) {
				input := validInput
				input.ChargeType = "invalid_type"
				return model.NewOffering(input)
			},
			assert: func(t *testing.T, o *model.Offering, err error) {
				assert.Nil(t, o)
				require.Error(t, err)
				assert.ErrorIs(t, err, model.ErrInvalidChargeType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offering, err := tt.arrange(t)
			tt.assert(t, offering, err)
		})
	}
}

func TestOffering_Activate(t *testing.T) {
	offering, err := model.NewOffering(model.OfferingInput{
		Name: "Test", Description: "Test",
		ChargeType: "one_time", Currency: "USD", Price: 100,
	})
	require.NoError(t, err)
	require.False(t, offering.IsActive)

	offering.Activate()

	assert.True(t, offering.IsActive)
	assert.NotNil(t, offering.UpdatedAt)
}

func TestOffering_Deactivate(t *testing.T) {
	offering, err := model.NewOffering(model.OfferingInput{
		Name: "Test", Description: "Test",
		ChargeType: "one_time", Currency: "USD", Price: 100, IsActive: true,
	})
	require.NoError(t, err)
	require.True(t, offering.IsActive)

	offering.Deactivate()

	assert.False(t, offering.IsActive)
	assert.NotNil(t, offering.UpdatedAt)
}

func TestChargeType_Valid(t *testing.T) {
	tests := []struct {
		chargeType model.ChargeType
		name       string
		valid      bool
	}{
		{name: "one_time should be valid", chargeType: model.OneTimeCharge, valid: true},
		{name: "subscription should be valid", chargeType: model.SubscriptionCharge, valid: true},
		{name: "empty string should be invalid", chargeType: model.ChargeType(""), valid: false},
		{name: "unknown should be invalid", chargeType: model.ChargeType("invalid"), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.chargeType.Valid())
		})
	}
}

func TestChargeType_String(t *testing.T) {
	assert.Equal(t, "one_time", model.OneTimeCharge.String())
	assert.Equal(t, "subscription", model.SubscriptionCharge.String())
}
