package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/http/handler"
	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/service"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestHandler_GetPayment(t *testing.T) {
	newRouter := func(h *handler.Handler) *chi.Mux {
		r := chi.NewRouter()
		r.Get("/v1/payments/{id}", h.GetPayment)
		return r
	}

	tests := []struct {
		setupMocks     func(*mocks.PaymentService)
		name           string
		path           string
		expectedStatus int
	}{
		{
			name: "success",
			path: "/v1/payments/payment-123",
			setupMocks: func(m *mocks.PaymentService) {
				m.On("GetPayment", mock.Anything, service.GetPaymentInput{PaymentID: "payment-123"}).
					Return(&model.Payment{Status: model.PaidPaymentStatus}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found error",
			path: "/v1/payments/unknown",
			setupMocks: func(m *mocks.PaymentService) {
				m.On("GetPayment", mock.Anything, service.GetPaymentInput{PaymentID: "unknown"}).
					Return(nil, errNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPayment := new(mocks.PaymentService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockPayment)
			}

			h := handler.New(new(mocks.CatalogService), mockPayment)
			router := newRouter(h)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp model.Payment
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, model.PaidPaymentStatus, resp.Status)
			}
			mockPayment.AssertExpectations(t)
		})
	}
}
