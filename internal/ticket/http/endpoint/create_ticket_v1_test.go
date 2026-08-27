package endpoint_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/http/endpoint"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	"github.com/charmingruby/lab/pkg/o11y"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
)

func TestMain(m *testing.M) {
	o11y.InitLogger()
	os.Exit(m.Run())
}

func TestCreateTicketV1(t *testing.T) {
	tests := []struct {
		body          any
		mockSetup     func(uc *mocks.MockCreateTicketUsecase)
		wantBodyCheck func(t *testing.T, body map[string]any)
		name          string
		wantStatus    int
	}{
		{
			name:       "invalid JSON body returns 400",
			body:       "not json",
			mockSetup:  func(uc *mocks.MockCreateTicketUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Contains(t, body["message"], "invalid payload")
			},
		},
		{
			name: "missing required fields returns 400",
			body: map[string]string{
				"title": "Test Ticket",
			},
			mockSetup:  func(uc *mocks.MockCreateTicketUsecase) {},
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Contains(t, body["message"], "invalid payload")
			},
		},
		{
			name: "usecase validation error returns 400",
			body: map[string]string{
				"title":       "Test Ticket",
				"description": "A description",
				"priority":    "invalid",
			},
			mockSetup: func(uc *mocks.MockCreateTicketUsecase) {
				uc.EXPECT().
					CreateTicket(mock.Anything, usecase.CreateTicketInput{
						Title:       "Test Ticket",
						Description: "A description",
						Priority:    "invalid",
					}).
					Return(usecase.CreateTicketOutput{}, customerr.Validation("invalid ticket priority"))
			},
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "invalid ticket priority", body["message"])
			},
		},
		{
			name: "usecase integration error returns 500",
			body: map[string]string{
				"title":       "Test Ticket",
				"description": "A description",
				"priority":    "low",
			},
			mockSetup: func(uc *mocks.MockCreateTicketUsecase) {
				uc.EXPECT().
					CreateTicket(mock.Anything, usecase.CreateTicketInput{
						Title:       "Test Ticket",
						Description: "A description",
						Priority:    "low",
					}).
					Return(usecase.CreateTicketOutput{}, customerr.Integration(errors.New("db error")))
			},
			wantStatus: http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "db error", body["message"])
			},
		},
		{
			name: "success returns 201 with ticket ID",
			body: map[string]string{
				"title":       "Test Ticket",
				"description": "A description",
				"priority":    "high",
			},
			mockSetup: func(uc *mocks.MockCreateTicketUsecase) {
				uc.EXPECT().
					CreateTicket(mock.Anything, usecase.CreateTicketInput{
						Title:       "Test Ticket",
						Description: "A description",
						Priority:    "high",
					}).
					Return(usecase.CreateTicketOutput{ID: "new-ticket-id"}, nil)
			},
			wantStatus: http.StatusCreated,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "new-ticket-id", body["id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := mocks.NewMockCreateTicketUsecase(t)
			tt.mockSetup(uc)

			ep := endpoint.New(uc, nil, nil, nil)

			var reqBody *bytes.Buffer
			switch v := tt.body.(type) {
			case string:
				reqBody = bytes.NewBufferString(v)
			default:
				b, _ := json.Marshal(v)
				reqBody = bytes.NewBuffer(b)
			}

			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/tickets", reqBody)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			ep.CreateTicketV1(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			if rec.Body.Len() > 0 {
				_ = json.Unmarshal(rec.Body.Bytes(), &body)
			}
			tt.wantBodyCheck(t, body)
		})
	}
}
