package natsbus

import (
	"context"
	"fmt"

	"sen1or/letslive/shared/pkg/natsconn"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func connect(ctx context.Context, url string) (*nats.Conn, jetstream.JetStream, error) {
	conn, err := natsconn.Connect(ctx, url)
	if err != nil {
		return nil, nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	return conn, js, nil
}
