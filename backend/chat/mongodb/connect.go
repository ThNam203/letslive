package mongodb

import (
	"context"
	"fmt"
	"time"

	"sen1or/letslive/shared/pkg/logger"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	maxConnectAttempts = 10
	initialRetryDelay  = 2 * time.Second
	maxRetryDelay      = 30 * time.Second
	pingTimeout        = 5 * time.Second
)

// Connect retries until Mongo answers a ping, because chat_db is often still
// starting when the service comes up alongside it in docker-compose.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("invalid mongo configuration: %w", err)
	}

	retryDelay := initialRetryDelay
	var lastErr error

	for attempt := 1; attempt <= maxConnectAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
		lastErr = client.Ping(pingCtx, readpref.Primary())
		cancel()
		if lastErr == nil {
			return client, nil
		}

		logger.Warnf(ctx, "failed to reach mongo (attempt %d/%d): %v - retrying in %v...",
			attempt, maxConnectAttempts, lastErr, retryDelay)

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = client.Disconnect(context.Background())
			return nil, fmt.Errorf("context cancelled while connecting to mongo: %w", ctx.Err())
		case <-timer.C:
		}

		retryDelay = min(retryDelay*2, maxRetryDelay)
	}

	_ = client.Disconnect(context.Background())
	return nil, fmt.Errorf("failed to reach mongo after %d attempts: %w", maxConnectAttempts, lastErr)
}
