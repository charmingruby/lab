package endpoint_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/ticket/usecase"
)

func TestGetTicketV1(t *testing.T) {
	tests := []struct {
		wantBodyCheck func(t *testing.T, body map[string]any, ticketID string)
		name          string
		wantStatus    int
		seed          bool
	}{
		{
			name:       "ticket not found returns 404",
			seed:       false,
			wantStatus: http.StatusNotFound,
			wantBodyCheck: func(t *testing.T, body map[string]any, ticketID string) {
				assert.Equal(t, "ticket not found", body["message"])
			},
		},
		{
			name:       "success returns ticket",
			seed:       true,
			wantStatus: http.StatusOK,
			wantBodyCheck: func(t *testing.T, body map[string]any, ticketID string) {
				assert.Equal(t, ticketID, body["id"])
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
			ep, uc := newTestEndpoint(t)

			ticketID := "nonexistent"
			if tt.seed {
				created, err := uc.CreateTicket(context.Background(), usecase.CreateTicketInput{
					Title:       "Test Ticket",
					Description: "A description",
					Priority:    "high",
				})
				require.NoError(t, err)

				require.NoError(t, uc.AssignTicket(context.Background(), usecase.AssignTicketInput{
					TicketID:   created.ID,
					AssigneeID: "user-456",
				}))

				ticketID = created.ID
			}

			rec := func() *httptest.ResponseRecorder {
				req := httptest.NewRequestWithContext(
					context.Background(),
					http.MethodGet,
					"/v1/tickets/"+ticketID,
					nil,
				)
				req.Header.Set("Content-Type", "application/json")

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", ticketID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				rec := httptest.NewRecorder()
				ep.GetTicketV1(rec, req)

				return rec
			}()

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			tt.wantBodyCheck(t, body, ticketID)
		})
	}
}
