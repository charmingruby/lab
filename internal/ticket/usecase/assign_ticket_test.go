package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestAssignTicket(t *testing.T) {
	ctx := context.Background()
	assigneeID := "user-456"

	tests := []struct {
		seed              func(t *testing.T, s testService) string
		name              string
		assigneeID        string
		wantStatus        model.TicketStatus
		errType           customerr.ErrorType
		wantNotifications int
		wantErr           bool
	}{
		{
			name:       "ticket not found returns not found error",
			assigneeID: assigneeID,
			seed: func(t *testing.T, s testService) string {
				return "missing"
			},
			wantErr: true,
			errType: customerr.TypeNotFound,
		},
		{
			name:       "invalid transition returns validation error",
			assigneeID: assigneeID,
			seed: func(t *testing.T, s testService) string {
				t.Helper()

				ticketID := createTicket(t, s.uc)

				ticket, err := s.repo.FindByID(ctx, ticketID)
				require.NoError(t, err)
				require.NoError(t, ticket.Resolve())
				require.NoError(t, s.repo.Update(ctx, ticket))

				return ticketID
			},
			wantErr: true,
			errType: customerr.TypeValidation,
		},
		{
			name:       "success assigns ticket and sends notification",
			assigneeID: assigneeID,
			seed: func(t *testing.T, s testService) string {
				t.Helper()

				return createTicket(t, s.uc)
			},
			wantErr:           false,
			wantStatus:        model.InProgressTicketStatus,
			wantNotifications: 1,
		},
		{
			name:       "notification failure does not return error",
			assigneeID: assigneeID,
			seed: func(t *testing.T, s testService) string {
				t.Helper()

				s.notifier.SetError(errors.New("notification service down"))

				return createTicket(t, s.uc)
			},
			wantErr:           false,
			wantStatus:        model.InProgressTicketStatus,
			wantNotifications: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestService(t)
			ticketID := tt.seed(t, s)

			err := s.uc.AssignTicket(ctx, usecase.AssignTicketInput{
				TicketID:   ticketID,
				AssigneeID: tt.assigneeID,
			})

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, customerr.IsType(err, tt.errType))
				return
			}

			require.NoError(t, err)

			ticket, err := s.uc.GetTicket(ctx, usecase.GetTicketInput{TicketID: ticketID})
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, ticket.Status)
			require.NotNil(t, ticket.AssigneeID)
			assert.Equal(t, tt.assigneeID, *ticket.AssigneeID)

			sent := s.notifier.Sent()
			assert.Len(t, sent, tt.wantNotifications)
			if tt.wantNotifications > 0 {
				assert.Equal(t, tt.assigneeID, sent[0].AssigneeID)
				assert.Equal(t, "you have been assigned to a ticket", sent[0].Message)
			}
		})
	}
}
