// Package memory holds in-memory fakes for outbound client ports.
// Fakes implement the port with real state — they are not mocks:
// no call expectations, just behavior the usecase can observe.
//
//	uc := usecase.New(repo, txManager, memory.NewNotifier())
//	_ = uc.AssignTicket(ctx, input)
//	require.Len(t, notifier.Sent(), 1)
package memory

import (
	"context"
	"sync"

	"github.com/charmingruby/lab/internal/ticket/client"
)

type Notifier struct {
	mu   sync.Mutex
	sent []client.SendNotificationInput
	err  error
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(_ context.Context, input client.SendNotificationInput) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.err != nil {
		return n.err
	}

	n.sent = append(n.sent, input)

	return nil
}

func (n *Notifier) SetError(err error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.err = err
}

func (n *Notifier) Sent() []client.SendNotificationInput {
	n.mu.Lock()
	defer n.mu.Unlock()

	out := make([]client.SendNotificationInput, len(n.sent))
	copy(out, n.sent)

	return out
}
