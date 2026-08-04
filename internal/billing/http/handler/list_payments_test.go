package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/http/handler"
	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestHandler_ListPayments(t *testing.T) {
	type response struct {
		Payments []model.Payment `json:"payments"`
		Total    int             `json:"total"`
	}

	tests := []struct {
		setupMocks     func(*mocks.PaymentService)
		name           string
		query          string
		expectedResp   response
		expectedStatus int
	}{
		{
			name:  "success with default page",
			query: "user_id=user-123",
			setupMocks: func(m *mocks.PaymentService) {
				m.On("ListPayments", mock.Anything, service.ListPaymentsInput{
					UserID: "user-123",
					Page:   1,
				}).Return(service.ListPaymentsOutput{
					Payments: []model.Payment{
						{UserID: "user-123", Status: model.PaidPaymentStatus},
					},
					Total: 1,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: response{
				Payments: []model.Payment{
					{UserID: "user-123", Status: model.PaidPaymentStatus},
				},
				Total: 1,
			},
		},
		{
			name:  "success with custom page",
			query: "user_id=user-123&page=2",
			setupMocks: func(m *mocks.PaymentService) {
				m.On("ListPayments", mock.Anything, service.ListPaymentsInput{
					UserID: "user-123",
					Page:   2,
				}).Return(service.ListPaymentsOutput{
					Payments: []model.Payment{},
					Total:    0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: response{
				Payments: []model.Payment{},
				Total:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPayment := new(mocks.PaymentService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockPayment)
			}

			h := handler.New(new(mocks.CatalogService), mockPayment)
			w := httptest.NewRecorder()

			req := testRequest(t, http.MethodGet, "/v1/payments?"+tt.query, "")

			h.ListPayments(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp response
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedResp.Total, resp.Total)
				assert.Len(t, resp.Payments, len(tt.expectedResp.Payments))
			}
			mockPayment.AssertExpectations(t)
		})
	}
}
