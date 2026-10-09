package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestGetTicket(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		errType customerr.ErrorType
		seed    bool
		wantErr bool
	}{
		{
			name:    "ticket not found returns not found error",
			seed:    false,
			wantErr: true,
			errType: customerr.TypeNotFound,
		},
		{
			name:    "success returns ticket",
			seed:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestService(t)

			ticketID := "missing"
			if tt.seed {
				ticketID = createTicket(t, s.uc)
			}

			got, err := s.uc.GetTicket(ctx, usecase.GetTicketInput{TicketID: ticketID})

			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, customerr.IsType(err, tt.errType))
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, ticketID, got.ID)
			assert.Equal(t, "Test Ticket", got.Title)
		})
	}
}
