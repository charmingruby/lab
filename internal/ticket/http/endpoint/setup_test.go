package endpoint_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/platform/logging"
	"github.com/charmingruby/lab/internal/ticket/client/memory"
	"github.com/charmingruby/lab/internal/ticket/http/endpoint"
	"github.com/charmingruby/lab/internal/ticket/repository"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	"github.com/charmingruby/lab/test/container"
)

func TestMain(m *testing.M) {
	logging.InitLogger()
	os.Exit(m.Run())
}

func newTestEndpoint(t *testing.T) (*endpoint.Endpoint, *usecase.Usecase) {
	t.Helper()

	db := container.StartPostgres(t)

	repo, err := repository.NewTicketRepository(db)
	require.NoError(t, err)

	uc := usecase.New(repo, repository.NewTransactionManager(db), memory.NewNotifier())

	return endpoint.New(uc), uc
}

func mustBody(t *testing.T, v any) []byte {
	t.Helper()

	if v == nil {
		return nil
	}

	if s, ok := v.(string); ok {
		return []byte(s)
	}

	b, err := json.Marshal(v)
	require.NoError(t, err)

	return b
}

func withRouteParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func serve(
	t *testing.T,
	handler func(w http.ResponseWriter, r *http.Request),
	method, target string,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, target, reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}

func serveWithRouteParam(
	t *testing.T,
	handler func(w http.ResponseWriter, r *http.Request),
	method, target, paramKey, paramVal string,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	req = withRouteParam(req, paramKey, paramVal)

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}
