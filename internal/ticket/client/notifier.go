package client

import (
	"context"
)

type Notifier interface {
	Send(ctx context.Context, input SendNotificationInput) error
}

type SendNotificationInput struct {
	AssigneeID string
	Message    string
}
