package natsconn

import (
	"context"
	"fmt"
	"time"

	"sen1or/letslive/shared/pkg/logger"

	"github.com/nats-io/nats.go"
)

const (
	maxConnectAttempts = 10
	initialRetryDelay  = 2 * time.Second
	maxRetryDelay      = 30 * time.Second
	reconnectWait      = 2 * time.Second
)

// Connect dials NATS with retry because the server is often not ready yet when
// a service starts alongside it in docker-compose. Once connected, nats.go
// reconnects forever and restores subscriptions on its own.
func Connect(ctx context.Context, url string) (*nats.Conn, error) {
	retryDelay := initialRetryDelay
	var lastErr error

	for attempt := 1; attempt <= maxConnectAttempts; attempt++ {
		conn, err := nats.Connect(url, nats.MaxReconnects(-1), nats.ReconnectWait(reconnectWait))
		if err == nil {
			return conn, nil
		}
		lastErr = err

		logger.Warnf(ctx, "failed to connect to nats at %s (attempt %d/%d): %v - retrying in %v...",
			url, attempt, maxConnectAttempts, err, retryDelay)

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("context cancelled while connecting to nats: %w", ctx.Err())
		case <-timer.C:
		}

		retryDelay = min(retryDelay*2, maxRetryDelay)
	}

	return nil, fmt.Errorf("failed to connect to nats after %d attempts: %w", maxConnectAttempts, lastErr)
}
