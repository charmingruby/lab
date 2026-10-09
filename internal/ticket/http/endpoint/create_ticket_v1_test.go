package endpoint_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestCreateTicketV1(t *testing.T) {
	tests := []struct {
		body          any
		wantBodyCheck func(t *testing.T, body map[string]any)
		name          string
		wantStatus    int
		wantPersisted bool
	}{
		{
			name:       "invalid JSON body returns 400",
			body:       "not json",
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
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Contains(t, body["message"], "invalid payload")
			},
		},
		{
			name: "invalid priority returns 400",
			body: map[string]string{
				"title":       "Test Ticket",
				"description": "A description",
				"priority":    "invalid",
			},
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "invalid ticket priority", body["message"])
			},
		},
		{
			name: "success returns 201 and persists the ticket",
			body: map[string]string{
				"title":       "Test Ticket",
				"description": "A description",
				"priority":    "high",
			},
			wantStatus: http.StatusCreated,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["id"])
			},
			wantPersisted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, uc := newTestEndpoint(t)

			var bodyBytes []byte
			if tt.body != nil {
				if s, ok := tt.body.(string); ok {
					bodyBytes = []byte(s)
				} else {
					var err error
					bodyBytes, err = json.Marshal(tt.body)
					require.NoError(t, err)
				}
			}

			var reader io.Reader
			if bodyBytes != nil {
				reader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/v1/tickets",
				reader,
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ep.CreateTicketV1(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			tt.wantBodyCheck(t, body)

			if tt.wantPersisted {
				id, ok := body["id"].(string)
				require.True(t, ok, "expected string id in response body")
				ticket, err := uc.GetTicket(context.Background(), usecase.GetTicketInput{
					TicketID: id,
				})
				require.NoError(t, err)
				assert.Equal(t, "Test Ticket", ticket.Title)
			}
		})
	}
}
