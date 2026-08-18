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

func TestEndpoint_CreateOfferingV1(t *testing.T) {
	type response struct {
		ID string `json:"id"`
	}

	tests := []struct {
		setupMocks     func(*mocks.CreateOfferingUsecase)
		name           string
		body           string
		expectedID     string
		expectedStatus int
	}{
		{
			name: "success",
			body: `{"name":"Premium Plan","description":"Premium plan","charge_type":"one_time","currency":"USD","price":2999,"is_active":true}`,
			setupMocks: func(m *mocks.CreateOfferingUsecase) {
				m.On("CreateOffering", mock.Anything, model.OfferingInput{
					Name: "Premium Plan", Description: "Premium plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999, IsActive: true,
				}).Return(usecase.CreateOfferingOutput{ID: "offering-123"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedID:     "offering-123",
		},
		{
			name:           "validation error",
			body:           `{"name":""}`,
			setupMocks:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "conflict error",
			body: `{"name":"Existing Plan","description":"Test","charge_type":"one_time","currency":"USD","price":100}`,
			setupMocks: func(m *mocks.CreateOfferingUsecase) {
				m.On("CreateOffering", mock.Anything, mock.Anything).
					Return(usecase.CreateOfferingOutput{}, errConflict)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCreateOffering := new(mocks.CreateOfferingUsecase)
			if tt.setupMocks != nil {
				tt.setupMocks(mockCreateOffering)
			}

			ep := endpoint.New(
				mockCreateOffering,
				new(mocks.CreatePaymentUsecase),
				new(mocks.GetPaymentUsecase),
				new(mocks.ListPaymentsUsecase),
			)
			w := httptest.NewRecorder()
			req := testRequest(t, http.MethodPost, "/v1/offerings/", tt.body)

			ep.CreateOfferingV1(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedID != "" {
				var resp response
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedID, resp.ID)
			}
			mockCreateOffering.AssertExpectations(t)
		})
	}
}
