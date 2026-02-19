package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suka712/api.sukaseven.com/internal/db/generated"
)

func Connect(ctx context.Context, connString string) (*gendb.Queries, *pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, nil, err
	}

	queries := gendb.New(pool)
	
	return queries, pool, nil
}
