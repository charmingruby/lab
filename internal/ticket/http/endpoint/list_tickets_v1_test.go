package endpoint_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/internal/shared/customerr"
	"github.com/charmingruby/lab/internal/ticket/http/endpoint"
	"github.com/charmingruby/lab/internal/ticket/model"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	mocks "github.com/charmingruby/lab/test/ticket/mocks"
)

func TestListTicketsV1(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tickets := []model.Ticket{
		{
			Title:       "Ticket 1",
			Description: "First ticket",
			Status:      model.OpenTicketStatus,
			Priority:    model.HighPriority,
		},
		{
			Title:       "Ticket 2",
			Description: "Second ticket",
			Status:      model.OpenTicketStatus,
			Priority:    model.LowPriority,
		},
	}
	tickets[0].ID = "ticket-1"
	tickets[0].CreatedAt = now
	tickets[1].ID = "ticket-2"
	tickets[1].CreatedAt = now

	tests := []struct {
		name          string
		queryParams   string
		setupMock     func(uc *mocks.MockListTicketsUsecase)
		wantStatus    int
		wantBodyCheck func(t *testing.T, body map[string]any)
	}{
		{
			name:        "missing status query param returns 500",
			queryParams: "",
			setupMock:   func(uc *mocks.MockListTicketsUsecase) {},
			wantStatus:  http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "Internal Server Error", body["message"])
			},
		},
		{
			name:        "integration error returns 500",
			queryParams: "?status=open&page=1&limit=25",
			setupMock: func(uc *mocks.MockListTicketsUsecase) {
				uc.EXPECT().
					ListTickets(mock.Anything, usecase.ListTicketsInput{
						Status: "open",
						Params: core.PaginationParams{Page: 1, Limit: 25},
					}).
					Return(usecase.ListTicketsOutput{}, customerr.Integration(errors.New("db error")))
			},
			wantStatus: http.StatusInternalServerError,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "db error", body["message"])
			},
		},
		{
			name:        "success returns tickets list",
			queryParams: "?status=open&page=1&limit=25",
			setupMock: func(uc *mocks.MockListTicketsUsecase) {
				uc.EXPECT().
					ListTickets(mock.Anything, usecase.ListTicketsInput{
						Status: "open",
						Params: core.PaginationParams{Page: 1, Limit: 25},
					}).
					Return(usecase.ListTicketsOutput{
						Tickets:    tickets,
						Page:       1,
						Limit:      25,
						Total:      2,
						TotalPages: 1,
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, float64(2), body["total"])
				assert.Equal(t, float64(1), body["page"])
				assert.Equal(t, float64(25), body["limit"])
				assert.Equal(t, float64(1), body["total_pages"])

				ticketList, ok := body["tickets"].([]any)
				assert.True(t, ok)
				assert.Len(t, ticketList, 2)

				first := ticketList[0].(map[string]any)
				assert.Equal(t, "ticket-1", first["id"])
				assert.Equal(t, "Ticket 1", first["title"])
				assert.Equal(t, "open", first["status"])
				assert.Equal(t, "high", first["priority"])
			},
		},
		{
			name:        "empty result returns empty list",
			queryParams: "?status=resolved&page=1&limit=25",
			setupMock: func(uc *mocks.MockListTicketsUsecase) {
				uc.EXPECT().
					ListTickets(mock.Anything, usecase.ListTicketsInput{
						Status: "resolved",
						Params: core.PaginationParams{Page: 1, Limit: 25},
					}).
					Return(usecase.ListTicketsOutput{
						Tickets:    []model.Ticket{},
						Page:       1,
						Limit:      25,
						Total:      0,
						TotalPages: 0,
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantBodyCheck: func(t *testing.T, body map[string]any) {
				assert.Equal(t, float64(0), body["total"])

				ticketList, ok := body["tickets"].([]any)
				assert.True(t, ok)
				assert.Empty(t, ticketList)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := mocks.NewMockListTicketsUsecase(t)
			tt.setupMock(uc)

			ep := endpoint.New(nil, nil, nil, uc)

			url := "/v1/tickets" + tt.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			ep.ListTicketsV1(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			if rec.Body.Len() > 0 {
				_ = json.Unmarshal(rec.Body.Bytes(), &body)
			}
			tt.wantBodyCheck(t, body)
		})
	}
}
