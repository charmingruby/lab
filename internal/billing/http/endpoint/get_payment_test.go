package endpoint_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/http/endpoint"
	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/test/billing/mocks"
)

func TestEndpoint_GetPaymentV1(t *testing.T) {
	newRouter := func(ep *endpoint.Endpoint) *chi.Mux {
		r := chi.NewRouter()
		r.Get("/v1/payments/{id}", ep.GetPaymentV1)
		return r
	}

	tests := []struct {
		setupMocks     func(*mocks.GetPaymentUsecase)
		name           string
		path           string
		expectedStatus int
	}{
		{
			name: "success",
			path: "/v1/payments/payment-123",
			setupMocks: func(m *mocks.GetPaymentUsecase) {
				m.On("GetPayment", mock.Anything, usecase.GetPaymentInput{PaymentID: "payment-123"}).
					Return(&model.Payment{Status: model.PaidPaymentStatus}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found error",
			path: "/v1/payments/unknown",
			setupMocks: func(m *mocks.GetPaymentUsecase) {
				m.On("GetPayment", mock.Anything, usecase.GetPaymentInput{PaymentID: "unknown"}).
					Return(nil, errNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetPayment := new(mocks.GetPaymentUsecase)
			if tt.setupMocks != nil {
				tt.setupMocks(mockGetPayment)
			}

			ep := endpoint.New(
				new(mocks.CreateOfferingUsecase),
				new(mocks.CreatePaymentUsecase),
				mockGetPayment,
				new(mocks.ListPaymentsUsecase),
			)
			router := newRouter(ep)

			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

				assert.Equal(t, model.PaidPaymentStatus, resp["status"])
			}

			mockGetPayment.AssertExpectations(t)
		})
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
