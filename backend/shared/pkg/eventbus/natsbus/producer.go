package natsbus

import (
	"context"
	"encoding/json"
	"fmt"

	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type natsProducer struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// NewProducer creates a new NATS JetStream-backed event producer.
func NewProducer(ctx context.Context, url string) (eventbus.Producer, error) {
	conn, js, err := connect(ctx, url)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "nats producer initialized, connected to %s", url)

	return &natsProducer{conn: conn, js: js}, nil
}

func (p *natsProducer) Publish(ctx context.Context, topic string, key string, event eventbus.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &nats.Msg{
		Subject: topic,
		Data:    data,
		Header: nats.Header{
			"Event-Type":   []string{event.Type},
			"Event-Source": []string{event.Source},
			"Event-Key":    []string{key},
		},
	}

	// dedupe on the event id within JetStream's default dedup window
	if _, err := p.js.PublishMsg(ctx, msg, jetstream.WithMsgID(event.ID)); err != nil {
		logger.Errorf(ctx, "failed to publish event %s to topic %s: %v", event.Type, topic, err)
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}

	logger.Debugf(ctx, "published event %s (id=%s) to topic %s with key %s", event.Type, event.ID, topic, key)
	return nil
}

func (p *natsProducer) Close() error {
	logger.Infof(context.TODO(), "closing nats producer...")
	p.conn.Close()
	return nil
}
