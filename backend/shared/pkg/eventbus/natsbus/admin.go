package natsbus

import (
	"context"
	"fmt"

	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type natsAdmin struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// NewAdmin creates a new NATS JetStream-backed admin for stream management.
func NewAdmin(ctx context.Context, url string) (eventbus.Admin, error) {
	conn, js, err := connect(ctx, url)
	if err != nil {
		return nil, err
	}

	return &natsAdmin{conn: conn, js: js}, nil
}

// EnsureTopics creates (or updates) one JetStream stream per topic, with the
// topic name as the stream's sole subject. NumPartitions has no JetStream
// equivalent and is ignored, matching the interface's documented contract.
func (a *natsAdmin) EnsureTopics(ctx context.Context, topics []eventbus.TopicConfig) error {
	for _, t := range topics {
		cfg := jetstream.StreamConfig{
			Name:     streamNameFor(t.Name),
			Subjects: []string{t.Name},
			Storage:  jetstream.FileStorage,
		}
		if t.ReplicationFactor > 0 {
			cfg.Replicas = t.ReplicationFactor
		}

		if _, err := a.js.CreateOrUpdateStream(ctx, cfg); err != nil {
			return fmt.Errorf("failed to ensure stream for topic %s: %w", t.Name, err)
		}

		logger.Infof(ctx, "ensured nats stream '%s' for topic '%s' (replicas=%d)", cfg.Name, t.Name, cfg.Replicas)
	}

	return nil
}

func (a *natsAdmin) Close() error {
	a.conn.Close()
	return nil
}
