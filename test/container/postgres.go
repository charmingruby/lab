// Package container spins real infrastructure for integration tests.
// Golden sources (postgres) always run as containers — no mocks, no fakes.
//
//	t.Run("creates and reads back", func(t *testing.T) {
//		db := container.StartPostgres(t)
//		repo, err := repository.NewTicketRepository(db)
//		...
//	})
//
// Migrations in db/migration run automatically. Set TEST_DATABASE_URL to
// reuse an external postgres (e.g. CI without docker) instead of a container.
package container

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	pg "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/charmingruby/lab/internal/platform/postgrex"
)

const (
	postgresImage  = "postgres:16-alpine"
	postgresUser   = "postgres"
	postgresPass   = "postgres"
	postgresDBName = "lab_test"
)

func StartPostgres(t *testing.T) *sqlx.DB {
	t.Helper()

	ctx := context.Background()

	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return connectExternal(t, ctx, url)
	}

	ctr, err := pg.Run(
		ctx,
		postgresImage,
		pg.WithUsername(postgresUser),
		pg.WithPassword(postgresPass),
		pg.WithDatabase(postgresDBName),
		pg.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := postgrex.Connect(ctx, connStr)
	require.NoError(t, err)

	runMigrations(t, db)

	t.Cleanup(func() {
		_ = db.Close()
		_ = ctr.Terminate(ctx)
	})

	return db
}

func connectExternal(t *testing.T, ctx context.Context, url string) *sqlx.DB {
	t.Helper()

	db, err := postgrex.Connect(ctx, url)
	require.NoError(t, err)

	runMigrations(t, db)

	_, err = db.Exec("TRUNCATE tickets")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Exec("TRUNCATE tickets")
		_ = db.Close()
	})

	return db
}

func runMigrations(t *testing.T, db *sqlx.DB) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(repoRoot(t), "db", "migration", "*.up.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, files, "no migration files found")
	sort.Strings(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		require.NoError(t, err)

		_, err = db.Exec(string(sql))
		require.NoError(t, err, "migration %s", filepath.Base(f))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "go.mod not found")
		dir = parent
	}
}
