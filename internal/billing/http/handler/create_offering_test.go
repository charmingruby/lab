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

func TestHandler_CreateOffering(t *testing.T) {
	type response struct {
		ID string `json:"id"`
	}

	tests := []struct {
		setupMocks     func(*mocks.CatalogService)
		name           string
		body           string
		expectedID     string
		expectedStatus int
	}{
		{
			name: "success",
			body: `{"name":"Premium Plan","description":"Premium plan","charge_type":"one_time","currency":"USD","price":2999,"is_active":true}`,
			setupMocks: func(m *mocks.CatalogService) {
				m.On("CreateOffering", mock.Anything, model.OfferingInput{
					Name: "Premium Plan", Description: "Premium plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999, IsActive: true,
				}).Return(service.CreateOfferingOutput{ID: "offering-123"}, nil)
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
			setupMocks: func(m *mocks.CatalogService) {
				m.On("CreateOffering", mock.Anything, mock.Anything).
					Return(service.CreateOfferingOutput{}, errConflict)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCatalog := new(mocks.CatalogService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockCatalog)
			}

			h := handler.New(mockCatalog, new(mocks.PaymentService))
			w := httptest.NewRecorder()
			req := testRequest(t, http.MethodPost, "/v1/offerings/", tt.body)

			h.CreateOffering(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedID != "" {
				var resp response
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedID, resp.ID)
			}
			mockCatalog.AssertExpectations(t)
		})
	}
}
