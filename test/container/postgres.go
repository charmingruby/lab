package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	runMigrations(t, connStr)

	db, err := postgrex.Connect(ctx, connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
		_ = ctr.Terminate(ctx)
	})

	return db
}

func connectExternal(t *testing.T, ctx context.Context, url string) *sqlx.DB {
	t.Helper()

	runMigrations(t, url)

	db, err := postgrex.Connect(ctx, url)
	require.NoError(t, err)

	resetDatabase(t, db)

	t.Cleanup(func() {
		resetDatabase(t, db)
		_ = db.Close()
	})

	return db
}

func runMigrations(t *testing.T, databaseURL string) {
	t.Helper()

	m, err := migrate.New(
		"file://"+filepath.Join(repoRoot(t), "db", "migration"),
		databaseURL,
	)
	require.NoError(t, err)

	defer func() {
		_, _ = m.Close()
	}()

	err = m.Up()
	require.True(t, err == nil || errors.Is(err, migrate.ErrNoChange), "migrate up: %v", err)
}

func resetDatabase(t *testing.T, db *sqlx.DB) {
	t.Helper()

	var tables []string
	err := db.Select(
		&tables,
		`SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename NOT IN ('schema_migrations')`,
	)
	require.NoError(t, err)

	if len(tables) == 0 {
		return
	}

	quoted := make([]string, len(tables))
	for i, table := range tables {
		quoted[i] = fmt.Sprintf(`"%s"`, table)
	}

	_, err = db.Exec(fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", strings.Join(quoted, ", ")))
	require.NoError(t, err)
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
