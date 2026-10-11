package endpoint_test

import (
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/platform/logging"
	"github.com/charmingruby/lab/internal/ticket/client/notifier"
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

	uc := usecase.New(repo, repository.NewTransactionManager(db), notifier.NewMemory())

	return endpoint.New(chi.NewRouter(), uc), uc
}
