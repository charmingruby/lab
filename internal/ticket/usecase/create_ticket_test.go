package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTicket(t *testing.T) {
	tests := []struct {
		name       string
		input      usecase.CreateTicketInput
		mockSetup  func(repo *mocks.MockTicketRepository)
		wantErr    bool
		errType    customerr.ErrorType
	}{
		{
			name: "invalid priority returns validation error",
			input: usecase.CreateTicketInput{
				Title:       "Test Ticket",
				Description: "A description",
				Priority:    "invalid",
			},
			mockSetup: func(repo *mocks.MockTicketRepository) {},
			wantErr:    true,
			errType:    customerr.TypeValidation,
		},
		{
			name: "repository create error returns integration error",
			input: usecase.CreateTicketInput{
				Title:       "Test Ticket",
				Description: "A description",
				Priority:    "low",
			},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(errors.New("db connection failed"))
			},
			wantErr: true,
			errType: customerr.TypeIntegration,
		},
		{
			name: "success creates ticket and returns ID",
			input: usecase.CreateTicketInput{
				Title:       "Test Ticket",
				Description: "A description",
				Priority:    "high",
			},
			mockSetup: func(repo *mocks.MockTicketRepository) {
				repo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTicketRepository(t)
			tt.mockSetup(repo)

			uc := usecase.NewCreateTicketUsecase(repo)

			got, err := uc.CreateTicket(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, customerr.IsType(err, tt.errType))
				assert.Empty(t, got.ID)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, got.ID)
		})
	}
}
