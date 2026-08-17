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
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestEndpoint_CreatePaymentV1(t *testing.T) {
	type response struct {
		ID string `json:"id"`
	}

	tests := []struct {
		setupMocks     func(*mocks.CreatePaymentUsecase)
		name           string
		body           string
		expectedID     string
		expectedStatus int
	}{
		{
			name: "success",
			body: `{"user_id":"user-123","offering_id":"offering-123","external_id":"pay_123","charged_amount":2999}`,
			setupMocks: func(m *mocks.CreatePaymentUsecase) {
				m.On("CreatePayment", mock.Anything, usecase.CreatePaymentInput{
					UserID:        "user-123",
					OfferingID:    "offering-123",
					ExternalID:    "pay_123",
					ChargedAmount: 2999,
				}).Return(usecase.CreatePaymentOutput{ID: "payment-123"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedID:     "payment-123",
		},
		{
			name:           "validation error",
			body:           `{"user_id":""}`,
			setupMocks:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "not found error",
			body: `{"user_id":"user-123","offering_id":"unknown","external_id":"pay_123","charged_amount":2999}`,
			setupMocks: func(m *mocks.CreatePaymentUsecase) {
				m.On("CreatePayment", mock.Anything, mock.Anything).
					Return(usecase.CreatePaymentOutput{}, errNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCreatePayment := new(mocks.CreatePaymentUsecase)
			if tt.setupMocks != nil {
				tt.setupMocks(mockCreatePayment)
			}

			h := endpoint.New(
				new(mocks.CreateOfferingUsecase),
				mockCreatePayment,
				new(mocks.GetPaymentUsecase),
				new(mocks.ListPaymentsUsecase),
			)
			w := httptest.NewRecorder()
			req := testRequest(t, http.MethodPost, "/v1/payments/", tt.body)

			h.CreatePaymentV1(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedID != "" {
				var resp response
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedID, resp.ID)
			}
			mockCreatePayment.AssertExpectations(t)
		})
	}
}
