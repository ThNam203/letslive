package natsbus

import (
	"context"
	"fmt"
	"time"

	"sen1or/letslive/shared/pkg/logger"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// connect dials the NATS server with the same retry/backoff shape used by
// kafkabus's admin (2s initial delay, doubling, capped at 30s, 10 attempts),
// since JetStream is often not ready yet when a service starts alongside it
// in docker-compose.
func connect(ctx context.Context, url string) (*nats.Conn, jetstream.JetStream, error) {
	retryDelay := 2 * time.Second
	maxRetries := 10

	var conn *nats.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = nats.Connect(url, nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
		if err == nil {
			break
		}

		logger.Warnf(ctx, "failed to connect to nats at %s (attempt %d/%d): %v - retrying in %v...",
			url, i+1, maxRetries, err, retryDelay)

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, nil, fmt.Errorf("context cancelled while connecting to nats: %w", ctx.Err())
		case <-timer.C:
		}

		retryDelay *= 2
		if retryDelay > 30*time.Second {
			retryDelay = 30 * time.Second
		}
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to nats after %d attempts: %w", maxRetries, err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	return conn, js, nil
}
