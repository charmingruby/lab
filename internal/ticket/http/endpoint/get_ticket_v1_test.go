package endpoint_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/http/endpoint"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
)

func TestGetTicketV1(t *testing.T) {
	assigneeID := "user-456"

	makeTicket := func() *model.Ticket {
		t := &model.Ticket{
			Title:       "Test Ticket",
			Description: "A description",
			Status:      model.InProgressTicketStatus,
			Priority:    model.HighPriority,
			AssigneeID:  &assigneeID,
		}
		t.ID = "ticket-123"
		t.CreatedAt = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		return t
	}

	tests := []struct {
		mockSetup     func(uc *mocks.MockGetTicketUsecase)
		wantBodyCheck func(t *testing.T, body map[string]any)
		name          string
		ticketID      string
		wantStatus    int
	}{
		{
			name:     "ticket not found returns 404",
			ticketID: "nonexistent",
			mockSetup: func(uc *mocks.MockGetTicketUsecase) {
				uc.EXPECT().
					GetTicket(mock.Anything, usecase.GetTicketInput{TicketID: "nonexistent"}).
					Return(nil, customerr.NotFound("ticket not found"))
			},
			wantStatus: http.StatusNotFound,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "ticket not found", body["message"])
			},
		},
		{
			name:     "integration error returns 500",
			ticketID: "ticket-123",
			mockSetup: func(uc *mocks.MockGetTicketUsecase) {
				uc.EXPECT().
					GetTicket(mock.Anything, usecase.GetTicketInput{TicketID: "ticket-123"}).
					Return(nil, customerr.Integration(errors.New("db error")))
			},
			wantStatus: http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "db error", body["message"])
			},
		},
		{
			name:     "success returns ticket",
			ticketID: "ticket-123",
			mockSetup: func(uc *mocks.MockGetTicketUsecase) {
				uc.EXPECT().
					GetTicket(mock.Anything, usecase.GetTicketInput{TicketID: "ticket-123"}).
					Return(makeTicket(), nil)
			},
			wantStatus: http.StatusOK,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "ticket-123", body["id"])
				assert.Equal(t, "Test Ticket", body["title"])
				assert.Equal(t, "A description", body["description"])
				assert.Equal(t, "in_progress", body["status"])
				assert.Equal(t, "high", body["priority"])
				assert.Equal(t, "user-456", body["assignee_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := mocks.NewMockGetTicketUsecase(t)
			tt.mockSetup(uc)

			ep := endpoint.New(nil, nil, uc, nil)

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/tickets/"+tt.ticketID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.ticketID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			ep.GetTicketV1(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			if rec.Body.Len() > 0 {
				_ = json.Unmarshal(rec.Body.Bytes(), &body)
			}
			tt.wantBodyCheck(t, body)
		})
	}
}
