package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	config.ConnConfig.Tracer = &tracer{}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return db, nil
}
