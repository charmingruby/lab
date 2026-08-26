package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTicket(t *testing.T) {
	ticketID := "ticket-123"

	existingTicket := &model.Ticket{
		Title:       "Test Ticket",
		Description: "A description",
		Status:      model.OpenTicketStatus,
		Priority:    model.MediumPriority,
	}
	existingTicket.ID = ticketID

	tests := []struct {
		name       string
		input      usecase.GetTicketInput
		mockSetup  func(repo *mocks.MockTicketRepository)
		want       *model.Ticket
		wantErr    bool
		errType    customerr.ErrorType
	}{
		{
			name:  "repository error returns integration error",
			input: usecase.GetTicketInput{TicketID: ticketID},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(nil, errors.New("db timeout"))
			},
			wantErr: true,
			errType: customerr.TypeIntegration,
		},
		{
			name:  "ticket not found returns not found error",
			input: usecase.GetTicketInput{TicketID: ticketID},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(nil, nil)
			},
			wantErr: true,
			errType: customerr.TypeNotFound,
		},
		{
			name:  "success returns ticket",
			input: usecase.GetTicketInput{TicketID: ticketID},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(existingTicket, nil)
			},
			want:    existingTicket,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTicketRepository(t)
			tt.mockSetup(repo)

			uc := usecase.NewGetTicketUsecase(repo)

			got, err := uc.GetTicket(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, customerr.IsType(err, tt.errType))
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
