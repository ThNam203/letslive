package utils

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(ctx context.Context, connectionString string) *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		logger.Panicf(ctx, "unable to parse database connection string: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	dbConn, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Panicf(ctx, "unable to connect to database: %v", err)
	}

	if err := dbConn.Ping(ctx); err != nil {
		logger.Panicf(ctx, "unable to ping database: %v", err)
	}

	return dbConn
}
