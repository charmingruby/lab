package postgrex

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(ctx context.Context, url string) (*sqlx.DB, error) {
	ctx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()

	db, err := sqlx.ConnectContext(ctx, "postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
