package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestCreateTicket(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		input   usecase.CreateTicketInput
		name    string
		errType customerr.ErrorType
		wantErr bool
	}{
		{
			name: "invalid priority returns validation error",
			input: usecase.CreateTicketInput{
				Title:       "Test Ticket",
				Description: "A description",
				Priority:    "invalid",
			},
			wantErr: true,
			errType: customerr.TypeValidation,
		},
		{
			name: "success creates ticket and reads back",
			input: usecase.CreateTicketInput{
				Title:       "Test Ticket",
				Description: "A description",
				Priority:    "high",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestService(t)

			got, err := s.uc.CreateTicket(ctx, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, customerr.IsType(err, tt.errType))
				assert.Empty(t, got.ID)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, got.ID)

			ticket, err := s.uc.GetTicket(ctx, usecase.GetTicketInput{TicketID: got.ID})
			require.NoError(t, err)
			assert.Equal(t, tt.input.Title, ticket.Title)
			assert.Equal(t, model.OpenTicketStatus, ticket.Status)
			assert.Equal(t, model.TicketPriority(tt.input.Priority), ticket.Priority)
		})
	}
}
