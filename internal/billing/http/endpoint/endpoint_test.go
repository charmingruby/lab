package endpoint_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/pkg/o11y"
)

func TestMain(m *testing.M) {
	o11y.InitLogger()
	os.Exit(m.Run())
}

var (
	errConflict = customerr.Conflict("conflict")
	errNotFound = customerr.NotFound("not found")
)

func testRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))

	req.Header.Set("Content-Type", "application/json")

	ctx := o11y.WithLogger(req.Context(), slog.New(slog.DiscardHandler))

	return req.WithContext(ctx)
}
