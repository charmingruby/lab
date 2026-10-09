package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestListTickets(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		params    core.PaginationParams
		seedCount int
		wantTotal int
		wantPages int
		wantLen   int
	}{
		{
			name:      "empty result returns zero total",
			seedCount: 0,
			params:    core.PaginationParams{Page: 1, Limit: 25},
			wantTotal: 0,
			wantPages: 0,
			wantLen:   0,
		},
		{
			name:      "success returns tickets with pagination",
			seedCount: 3,
			params:    core.PaginationParams{Page: 1, Limit: 2},
			wantTotal: 3,
			wantPages: 2,
			wantLen:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestService(t)

			for i := range tt.seedCount {
				_, err := s.uc.CreateTicket(ctx, usecase.CreateTicketInput{
					Title:       fmt.Sprintf("Ticket %d", i+1),
					Description: "A description",
					Priority:    "low",
				})
				require.NoError(t, err)
			}

			got, err := s.uc.ListTickets(ctx, usecase.ListTicketsInput{
				Status: "open",
				Params: tt.params,
			})
			require.NoError(t, err)
			assert.Len(t, got.Tickets, tt.wantLen)
			assert.Equal(t, tt.wantTotal, got.Total)
			assert.Equal(t, tt.wantPages, got.TotalPages)
			assert.Equal(t, tt.params.Page, got.Page)
			assert.Equal(t, tt.params.Limit, got.Limit)
		})
	}
}
