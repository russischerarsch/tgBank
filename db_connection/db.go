package dbconnection

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateConnection(ctx context.Context) (*pgxpool.Pool, error) {
	// connStr := os.Getenv("DB_CONN")
	// if connStr == "" {
	// 	return nil, errors.New("connection url is empty")
	// }
	connStr := "postgres://a1111:secret@localhost:5432/postgres?sslmode=disable"
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
