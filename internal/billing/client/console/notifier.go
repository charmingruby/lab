package console

import (
	"context"

	"github.com/charmingruby/new/internal/billing/client"
	"github.com/charmingruby/new/pkg/o11y"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(ctx context.Context, input client.SendNotificationInput) error {
	o11y.LoggerFromContext(ctx).Info("notification sent",
		"user_id", input.UserID,
		"message", input.Message,
	)

	return nil
}
