package model

import (
	"errors"

	"github.com/charmingruby/new/internal/shared/core"
)

type ChargeType string

const (
	OneTimeCharge      ChargeType = "one_time"
	SubscriptionCharge ChargeType = "subscription"
)

var (
	ErrInvalidChargeType = errors.New("invalid charge type")
)

type Offering struct {
	core.Model  `           db:",inline"`
	Name        string     `db:"name"        json:"name"`
	Description string     `db:"description" json:"description"`
	ChargeType  ChargeType `db:"charge_type" json:"charge_type"`
	Currency    string     `db:"currency"    json:"currency"`
	Price       int        `db:"price"       json:"price"`
	IsActive    bool       `db:"is_active"   json:"is_active"`
}

type OfferingInput struct {
	Name        string
	Description string
	ChargeType  string
	Currency    string
	Price       int
	IsActive    bool
}

func NewOffering(input OfferingInput) (*Offering, error) {
	chargeType := ChargeType(input.ChargeType)

	if !chargeType.Valid() {
		return nil, ErrInvalidChargeType
	}

	return &Offering{
		Model:       core.NewModel(),
		Name:        input.Name,
		Description: input.Description,
		ChargeType:  chargeType,
		Price:       input.Price,
		Currency:    input.Currency,
		IsActive:    input.IsActive,
	}, nil
}

func (o *Offering) Activate() {
	o.Touch(func(m *core.Model) {
		o.IsActive = true
	})
}

func (o *Offering) Deactivate() {
	o.Touch(func(m *core.Model) {
		o.IsActive = false
	})
}

func (k ChargeType) Valid() bool {
	switch k {
	case OneTimeCharge, SubscriptionCharge:
		return true
	default:
		return false
	}
}

func (k ChargeType) String() string {
	return string(k)
}
