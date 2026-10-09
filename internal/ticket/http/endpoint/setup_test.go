package endpoint_test

import (
	"os"
	"testing"

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
