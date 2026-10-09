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

func TestListTicketsV1(t *testing.T) {
	tests := []struct {
		wantBodyCheck func(t *testing.T, body map[string]any)
		name          string
		query         string
		wantStatus    int
		seedCount     int
	}{
		{
			name:       "missing status query param returns 500",
			query:      "",
			seedCount:  0,
			wantStatus: http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "Internal Server Error", body["message"])
			},
		},
		{
			name:       "success returns tickets list",
			query:      "?status=open&page=1&limit=25",
			seedCount:  2,
			wantStatus: http.StatusOK,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.InDelta(t, 2, body["total"], 0.001)
				assert.InDelta(t, 1, body["page"], 0.001)
				assert.InDelta(t, 25, body["limit"], 0.001)
				assert.InDelta(t, 1, body["total_pages"], 0.001)

				ticketList, ok := body["tickets"].([]any)
				require.True(t, ok)
				require.Len(t, ticketList, 2)

				first, ok := ticketList[0].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "open", first["status"])
			},
		},
		{
			name:       "empty result returns empty list",
			query:      "?status=resolved&page=1&limit=25",
			seedCount:  0,
			wantStatus: http.StatusOK,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.InDelta(t, 0, body["total"], 0.001)

				ticketList, ok := body["tickets"].([]any)
				require.True(t, ok)
				assert.Empty(t, ticketList)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, uc := newTestEndpoint(t)

			for range tt.seedCount {
				_, err := uc.CreateTicket(context.Background(), usecase.CreateTicketInput{
					Title:       "Ticket",
					Description: "A description",
					Priority:    "low",
				})
				require.NoError(t, err)
			}

			rec := serve(t, ep.ListTicketsV1, http.MethodGet, "/v1/tickets"+tt.query, nil)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			tt.wantBodyCheck(t, body)
		})
	}
}
