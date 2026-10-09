package usecase_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/charmingruby/lab/internal/platform/logging"
	"github.com/charmingruby/lab/internal/ticket/client/memory"
	"github.com/charmingruby/lab/internal/ticket/repository"
	"github.com/charmingruby/lab/internal/ticket/usecase"
	"github.com/charmingruby/lab/test/container"
)

func TestMain(m *testing.M) {
	logging.InitLogger()
	os.Exit(m.Run())
}

type testService struct {
	uc       *usecase.Usecase
	repo     *repository.TicketRepository
	notifier *memory.Notifier
}

func newTestService(t *testing.T) testService {
	t.Helper()

	db := container.StartPostgres(t)

	repo, err := repository.NewTicketRepository(db)
	require.NoError(t, err)

	notifier := memory.NewNotifier()

	return testService{
		uc:       usecase.New(repo, repository.NewTransactionManager(db), notifier),
		repo:     repo,
		notifier: notifier,
	}
}

func createTicket(t *testing.T, uc *usecase.Usecase) string {
	t.Helper()

	output, err := uc.CreateTicket(context.Background(), usecase.CreateTicketInput{
		Title:       "Test Ticket",
		Description: "A description",
		Priority:    "low",
	})
	require.NoError(t, err)

	return output.ID
}
