package notifier

import (
	"context"

	"github.com/charmingruby/lab/internal/platform/logging"
	"github.com/charmingruby/lab/internal/ticket/client"
)

type Console struct{}

func NewConsole() *Console {
	return &Console{}
}

func (n *Console) Send(ctx context.Context, input client.SendNotificationInput) error {
	logging.LoggerFromContext(ctx).Info("notification sent",
		"assignee_id", input.AssigneeID,
		"message", input.Message,
	)

	return nil
}
