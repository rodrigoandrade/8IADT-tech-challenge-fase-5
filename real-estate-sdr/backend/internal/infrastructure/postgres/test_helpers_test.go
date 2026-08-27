package postgres

import (
	"context"
	"os"
	generated "real-estate-sdr/internal/infrastructure/postgres/generated"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not configured")
	}

	pool, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"failed to create PostgreSQL pool: %v",
			err,
		)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()

		t.Fatalf(
			"failed to ping PostgreSQL: %v",
			err,
		)
	}

	return pool
}

func newQueries(pool *pgxpool.Pool) *generated.Queries {
	return generated.New(pool)
}
