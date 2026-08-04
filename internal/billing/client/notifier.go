package client

import (
	"context"

	"github.com/charmingruby/new/internal/billing/model"
)

type NotificationClient interface {
	Send(ctx context.Context, input SendNotificationInput) error
}

type SendNotificationInput struct {
	UserID  string
	Message string
}

type PaymentReader interface {
	GetPayment(ctx context.Context, paymentID string) (*model.Payment, error)
}
