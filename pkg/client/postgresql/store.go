package postgresql

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

func NewClient(cfg *Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
