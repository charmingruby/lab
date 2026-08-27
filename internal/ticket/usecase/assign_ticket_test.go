package usecase_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/client"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/repository"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	"github.com/charmingruby/lab/pkg/o11y"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
)

func TestMain(m *testing.M) {
	o11y.InitLogger()
	os.Exit(m.Run())
}

type fakeTxManager struct {
	txErr error
	repo  repository.TicketRepository
}

func (f *fakeTxManager) Transact(fn func(tx repository.Transaction) error) error {
	if f.txErr != nil {
		return f.txErr
	}
	return fn(repository.Transaction{TicketRepo: f.repo})
}

func newOpenTicket(id string) *model.Ticket {
	t := &model.Ticket{
		Title:       "Test Ticket",
		Description: "A description",
		Status:      model.OpenTicketStatus,
		Priority:    model.MediumPriority,
	}
	t.ID = id
	return t
}

func newResolvedTicket(id string) *model.Ticket {
	t := &model.Ticket{
		Title:       "Resolved Ticket",
		Description: "Already resolved",
		Status:      model.ResolvedTicketStatus,
		Priority:    model.LowPriority,
	}
	t.ID = id
	return t
}

func TestAssignTicket(t *testing.T) {
	ticketID := "ticket-123"
	assigneeID := "user-456"

	tests := []struct {
		setupMock func(t *testing.T) *fakeTxManager
		notifier  func(t *testing.T) *mocks.MockNotificationClient
		input     usecase.AssignTicketInput
		name      string
		errType   customerr.ErrorType
		wantErr   bool
	}{
		{
			name:  "transaction error returns error",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				return &fakeTxManager{txErr: customerr.Integration(errors.New("tx failed"))}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				return mocks.NewMockNotificationClient(t)
			},
			wantErr: true,
			errType: customerr.TypeIntegration,
		},
		{
			name:  "ticket not found returns not found error",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				repo := mocks.NewMockTicketRepository(t)
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(nil, nil)
				return &fakeTxManager{repo: repo}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				return mocks.NewMockNotificationClient(t)
			},
			wantErr: true,
		},
		{
			name:  "find by id error returns integration error",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				repo := mocks.NewMockTicketRepository(t)
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(nil, errors.New("db error"))
				return &fakeTxManager{repo: repo}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				return mocks.NewMockNotificationClient(t)
			},
			wantErr: true,
			errType: customerr.TypeIntegration,
		},
		{
			name:  "invalid transition returns validation error",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				repo := mocks.NewMockTicketRepository(t)
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(newResolvedTicket(ticketID), nil)
				return &fakeTxManager{repo: repo}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				return mocks.NewMockNotificationClient(t)
			},
			wantErr: true,
			errType: customerr.TypeValidation,
		},
		{
			name:  "update error returns integration error",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				repo := mocks.NewMockTicketRepository(t)
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(newOpenTicket(ticketID), nil)
				repo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(errors.New("update failed"))
				return &fakeTxManager{repo: repo}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				return mocks.NewMockNotificationClient(t)
			},
			wantErr: true,
			errType: customerr.TypeIntegration,
		},
		{
			name:  "success assigns ticket and sends notification",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				repo := mocks.NewMockTicketRepository(t)
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(newOpenTicket(ticketID), nil)
				repo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil)
				return &fakeTxManager{repo: repo}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				n := mocks.NewMockNotificationClient(t)
				n.EXPECT().
					Send(mock.Anything, client.SendNotificationInput{
						AssigneeID: assigneeID,
						Message:    "you have been assigned to a ticket",
					}).
					Return(nil)
				return n
			},
			wantErr: false,
		},
		{
			name:  "notification failure does not return error",
			input: usecase.AssignTicketInput{TicketID: ticketID, AssigneeID: assigneeID},
			setupMock: func(t *testing.T) *fakeTxManager {
				repo := mocks.NewMockTicketRepository(t)
				repo.EXPECT().
					FindByID(mock.Anything, ticketID).
					Return(newOpenTicket(ticketID), nil)
				repo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(nil)
				return &fakeTxManager{repo: repo}
			},
			notifier: func(t *testing.T) *mocks.MockNotificationClient {
				n := mocks.NewMockNotificationClient(t)
				n.EXPECT().
					Send(mock.Anything, client.SendNotificationInput{
						AssigneeID: assigneeID,
						Message:    "you have been assigned to a ticket",
					}).
					Return(errors.New("notification service down"))
				return n
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txMgr := tt.setupMock(t)
			notifier := tt.notifier(t)

			uc := usecase.NewAssignTicketUsecase(txMgr, notifier)

			err := uc.AssignTicket(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != "" {
					assert.True(
						t,
						customerr.IsType(err, tt.errType),
						"expected errType %s, got: %v (type: %T)",
						tt.errType,
						err,
						err,
					)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}
