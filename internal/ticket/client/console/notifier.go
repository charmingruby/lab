package console

import (
	"context"

	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/pkg/o11y"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(ctx context.Context, input client.SendNotificationInput) error {
	o11y.LoggerFromContext(ctx).Info("notification sent",
		"assignee_id", input.AssigneeID,
		"message", input.Message,
	)

	return nil
}
