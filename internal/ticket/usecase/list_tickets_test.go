package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListTickets(t *testing.T) {
	tickets := []model.Ticket{
		{
			Title:       "Ticket 1",
			Description: "First ticket",
			Status:      model.OpenTicketStatus,
			Priority:    model.HighPriority,
		},
		{
			Title:       "Ticket 2",
			Description: "Second ticket",
			Status:      model.OpenTicketStatus,
			Priority:    model.LowPriority,
		},
	}
	tickets[0].ID = "ticket-1"
	tickets[1].ID = "ticket-2"

	defaultParams := core.PaginationParams{Page: 1, Limit: core.DefaultPageSize}

	tests := []struct {
		name       string
		input      usecase.ListTicketsInput
		mockSetup  func(repo *mocks.MockTicketRepository)
		want       usecase.ListTicketsOutput
		wantErr    bool
		errType    customerr.ErrorType
	}{
		{
			name:  "repository error returns integration error",
			input: usecase.ListTicketsInput{Status: "open", Params: defaultParams},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					ListByStatus(mock.Anything, "open", defaultParams).
					Return(nil, 0, errors.New("db connection refused"))
			},
			wantErr: true,
			errType: customerr.TypeIntegration,
		},
		{
			name:  "success returns tickets and total",
			input: usecase.ListTicketsInput{Status: "open", Params: defaultParams},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					ListByStatus(mock.Anything, "open", defaultParams).
					Return(tickets, 2, nil)
			},
			want: usecase.ListTicketsOutput{
				Tickets:    tickets,
				Page:       1,
				Limit:      core.DefaultPageSize,
				Total:      2,
				TotalPages: 1,
			},
			wantErr: false,
		},
		{
			name:  "empty result returns empty slice",
			input: usecase.ListTicketsInput{Status: "resolved", Params: defaultParams},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					ListByStatus(mock.Anything, "resolved", defaultParams).
					Return([]model.Ticket{}, 0, nil)
			},
			want: usecase.ListTicketsOutput{
				Tickets:    []model.Ticket{},
				Page:       1,
				Limit:      core.DefaultPageSize,
				Total:      0,
				TotalPages: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTicketRepository(t)
			tt.mockSetup(repo)

			uc := usecase.NewListTicketsUsecase(repo)

			got, err := uc.ListTickets(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, customerr.IsType(err, tt.errType))
				assert.Equal(t, usecase.ListTicketsOutput{}, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
