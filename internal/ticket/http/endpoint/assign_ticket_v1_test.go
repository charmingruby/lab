package endpoint_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestAssignTicketV1(t *testing.T) {
	tests := []struct {
		body          any
		wantBodyCheck func(t *testing.T, body map[string]any)
		name          string
		ticketID      string
		wantStatus    int
		seed          bool
		wantAssigned  bool
	}{
		{
			name:       "invalid JSON body returns 400",
			ticketID:   "ticket-123",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Contains(t, body["message"], "invalid payload")
			},
		},
		{
			name:       "missing assignee_id returns 400",
			ticketID:   "ticket-123",
			body:       map[string]string{},
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
			wantStatus: http.StatusNotFound,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "ticket not found", body["message"])
			},
		},
		{
			name: "success assigns and returns 204",
			body: map[string]string{
				"assignee_id": "user-456",
			},
			seed:       true,
			wantStatus: http.StatusNoContent,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Empty(t, body)
			},
			wantAssigned: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, uc := newTestEndpoint(t)

			ticketID := tt.ticketID
			if tt.seed {
				created, err := uc.CreateTicket(context.Background(), usecase.CreateTicketInput{
					Title:       "Test Ticket",
					Description: "A description",
					Priority:    "low",
				})
				require.NoError(t, err)
				ticketID = created.ID
			}

			rec := func() *httptest.ResponseRecorder {
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
					http.MethodPatch,
					"/v1/tickets/"+ticketID+"/assign",
					reader,
				)
				req.Header.Set("Content-Type", "application/json")

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", ticketID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				rec := httptest.NewRecorder()
				ep.AssignTicketV1(rec, req)

				return rec
			}()

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			if rec.Body.Len() > 0 {
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			}
			tt.wantBodyCheck(t, body)

			if tt.wantAssigned {
				ticket, err := uc.GetTicket(context.Background(), usecase.GetTicketInput{TicketID: ticketID})
				require.NoError(t, err)
				assert.Equal(t, "in_progress", string(ticket.Status))
			}
		})
	}
}
