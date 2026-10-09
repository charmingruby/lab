package endpoint_test

import (
	"context"
	"encoding/json"
	"net/http"
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

			rec := serve(t, ep.CreateTicketV1, http.MethodPost, "/v1/tickets", mustBody(t, tt.body))

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			tt.wantBodyCheck(t, body)

			if tt.wantPersisted {
				ticket, err := uc.GetTicket(context.Background(), usecase.GetTicketInput{
					TicketID: body["id"].(string),
				})
				require.NoError(t, err)
				assert.Equal(t, "Test Ticket", ticket.Title)
			}
		})
	}
}
