package console

import (
	"context"

	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/internal/platform/logging"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(ctx context.Context, input client.SendNotificationInput) error {
	logging.LoggerFromContext(ctx).Info("notification sent",
		"assignee_id", input.AssigneeID,
		"message", input.Message,
	)

	return nil
}
