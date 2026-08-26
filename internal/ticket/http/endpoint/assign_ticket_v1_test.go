package endpoint_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/http/endpoint"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
)

func TestAssignTicketV1(t *testing.T) {
	tests := []struct {
		name          string
		ticketID      string
		body          any
		mockSetup     func(uc *mocks.MockAssignTicketUsecase)
		wantStatus    int
		wantBodyCheck func(t *testing.T, body map[string]any)
	}{
		{
			name:     "invalid JSON body returns 400",
			ticketID: "ticket-123",
			body:     "not json",
			mockSetup: func(uc *mocks.MockAssignTicketUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Contains(t, body["message"], "invalid payload")
			},
		},
		{
			name:     "missing assignee_id returns 400",
			ticketID: "ticket-123",
			body:     map[string]string{},
			mockSetup: func(uc *mocks.MockAssignTicketUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Contains(t, body["message"], "invalid payload")
			},
		},
		{
			name:     "ticket not found returns 404",
			ticketID: "nonexistent",
			body: map[string]string{
				"assignee_id": "user-456",
			},
			mockSetup: func(uc *mocks.MockAssignTicketUsecase) {
				uc.EXPECT().
					AssignTicket(mock.Anything, usecase.AssignTicketInput{
						TicketID:   "nonexistent",
						AssigneeID: "user-456",
					}).
					Return(customerr.NotFound("ticket not found"))
			},
			wantStatus: http.StatusNotFound,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "ticket not found", body["message"])
			},
		},
		{
			name:     "integration error returns 500",
			ticketID: "ticket-123",
			body: map[string]string{
				"assignee_id": "user-456",
			},
			mockSetup: func(uc *mocks.MockAssignTicketUsecase) {
				uc.EXPECT().
					AssignTicket(mock.Anything, usecase.AssignTicketInput{
						TicketID:   "ticket-123",
						AssigneeID: "user-456",
					}).
					Return(customerr.Integration(errors.New("update failed")))
			},
			wantStatus: http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "update failed", body["message"])
			},
		},
		{
			name:     "success returns 204",
			ticketID: "ticket-123",
			body: map[string]string{
				"assignee_id": "user-456",
			},
			mockSetup: func(uc *mocks.MockAssignTicketUsecase) {
				uc.EXPECT().
					AssignTicket(mock.Anything, usecase.AssignTicketInput{
						TicketID:   "ticket-123",
						AssigneeID: "user-456",
					}).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Empty(t, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := mocks.NewMockAssignTicketUsecase(t)
			tt.mockSetup(uc)

			ep := endpoint.New(nil, uc, nil, nil)

			var reqBody *bytes.Buffer
			switch v := tt.body.(type) {
			case string:
				reqBody = bytes.NewBufferString(v)
			default:
				b, _ := json.Marshal(v)
				reqBody = bytes.NewBuffer(b)
			}

			req := httptest.NewRequest(http.MethodPatch, "/v1/tickets/"+tt.ticketID+"/assign", reqBody)
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.ticketID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			ep.AssignTicketV1(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			if rec.Body.Len() > 0 {
				_ = json.Unmarshal(rec.Body.Bytes(), &body)
			}
			tt.wantBodyCheck(t, body)
		})
	}
}
