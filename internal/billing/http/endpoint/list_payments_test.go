package endpoint_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/http/endpoint"
	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestEndpoint_ListPaymentsV1(t *testing.T) {
	type response struct {
		Payments []map[string]any `json:"payments"`
		Total    int              `json:"total"`
	}

	tests := []struct {
		setupMocks     func(*mocks.ListPaymentsUsecase)
		name           string
		query          string
		expectedCount  int
		expectedStatus int
		expectedTotal  int
	}{
		{
			name:  "success with default page",
			query: "user_id=user-123",
			setupMocks: func(m *mocks.ListPaymentsUsecase) {
				m.On("ListPayments", mock.Anything, usecase.ListPaymentsInput{
					UserID: "user-123",
					Page:   1,
				}).Return(usecase.ListPaymentsOutput{
					Payments: []model.Payment{
						{UserID: "user-123", Status: model.PaidPaymentStatus},
					},
					Total: 1,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedTotal:  1,
		},
		{
			name:  "success with custom page",
			query: "user_id=user-123&page=2",
			setupMocks: func(m *mocks.ListPaymentsUsecase) {
				m.On("ListPayments", mock.Anything, usecase.ListPaymentsInput{
					UserID: "user-123",
					Page:   2,
				}).Return(usecase.ListPaymentsOutput{
					Payments: []model.Payment{},
					Total:    0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockListPayments := new(mocks.ListPaymentsUsecase)
			if tt.setupMocks != nil {
				tt.setupMocks(mockListPayments)
			}

			ep := endpoint.New(
				new(mocks.CreateOfferingUsecase),
				new(mocks.CreatePaymentUsecase),
				new(mocks.GetPaymentUsecase),
				mockListPayments,
			)
			w := httptest.NewRecorder()

			req := testRequest(t, http.MethodGet, "/v1/payments?"+tt.query, "")

			ep.ListPaymentsV1(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp response
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedTotal, resp.Total)
				assert.Len(t, resp.Payments, tt.expectedCount)

				if len(resp.Payments) > 0 {
					payment := resp.Payments[0]
					assert.NotContains(t, payment, "deleted_at")
					assert.NotContains(t, payment, "updated_at")
				}
			}
			mockListPayments.AssertExpectations(t)
		})
	}
}
