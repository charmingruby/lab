package notifier

import (
	"context"
	"sync"

	"github.com/charmingruby/lab/internal/ticket/client"
)

type Memory struct {
	err  error
	sent []client.SendNotificationInput
	mu   sync.Mutex
}

func NewMemory() *Memory {
	return &Memory{}
}

func (n *Memory) Send(_ context.Context, input client.SendNotificationInput) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.err != nil {
		return n.err
	}

	n.sent = append(n.sent, input)

	return nil
}

func (n *Memory) SetError(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.err = err
}

func (n *Memory) Sent() []client.SendNotificationInput {
	n.mu.Lock()
	defer n.mu.Unlock()

	out := make([]client.SendNotificationInput, len(n.sent))
	copy(out, n.sent)

	return out
}
