package model

import (
	"errors"
	"fmt"

	"github.com/charmingruby/new/internal/shared/core"
)

type PaymentStatus = string

const (
	PendingPaymentStatus PaymentStatus = "pending"
	PaidPaymentStatus    PaymentStatus = "paid"
	FailedPaymentStatus  PaymentStatus = "failed"
)

var ErrInvalidPaymentTransition = errors.New("invalid payment status transition")

type Payment struct {
	core.Model    `              db:",inline"`
	UserID        string        `db:"user_id"        json:"user_id"`
	OfferingID    string        `db:"offering_id"    json:"offering_id"`
	Status        PaymentStatus `db:"status"         json:"status"`
	ExternalID    string        `db:"external_id"    json:"external_id"`
	ChargedAmount int           `db:"charged_amount" json:"charged_amount"`
}

type PaymentInput struct {
	UserID        string
	OfferingID    string
	ExternalID    string
	ChargedAmount int
}

func NewPayment(input PaymentInput) *Payment {
	return &Payment{
		Model:         core.NewModel(),
		UserID:        input.UserID,
		OfferingID:    input.OfferingID,
		Status:        PendingPaymentStatus,
		ChargedAmount: input.ChargedAmount,
		ExternalID:    input.ExternalID,
	}
}

func (p *Payment) MarkAsPaid() error {
	return p.transitionTo(PaidPaymentStatus)
}

func (p *Payment) MarkAsFailed() error {
	return p.transitionTo(FailedPaymentStatus)
}

func (p *Payment) transitionTo(status PaymentStatus) error {
	if p.Status != PendingPaymentStatus {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidPaymentTransition, p.Status, status)
	}

	p.Touch(func(m *core.Model) {
		p.Status = status
	})

	return nil
}
