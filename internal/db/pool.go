package db

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateConnectionPool(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("ERROR=%s. CONNECTION_STRING=%s", err.Error(), connectionString)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ERROR=%s. CONNECTION_STRING=%s", err.Error(), connectionString)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ERROR=%s. CONNECTION_STRING=%s", err.Error(), connectionString)
	}

	return pool, nil
}
