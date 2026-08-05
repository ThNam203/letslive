package natsbus

import (
	"context"
	"encoding/json"
	"sync"

	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type natsConsumer struct {
	conn    *nats.Conn
	js      jetstream.JetStream
	groupID string
}

// NewConsumer creates a new NATS JetStream-backed event consumer for the given consumer group.
func NewConsumer(ctx context.Context, url string, groupID string) (eventbus.Consumer, error) {
	conn, js, err := connect(ctx, url)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "nats consumer initialized for group '%s', connected to %s", groupID, url)

	return &natsConsumer{conn: conn, js: js, groupID: groupID}, nil
}

// Subscribe starts one durable JetStream consumer per topic (each bound to
// the stream created by Admin.EnsureTopics for that topic) and blocks until
// ctx is cancelled. The consumer group id becomes the durable name, so
// multiple instances of the same service share progress the way a Kafka
// consumer group does.
func (c *natsConsumer) Subscribe(ctx context.Context, topics []string, handler eventbus.EventHandler) error {
	logger.Infof(ctx, "subscribing to topics %v as group '%s'", topics, c.groupID)

	var wg sync.WaitGroup
	for _, topic := range topics {
		wg.Add(1)
		go func(topic string) {
			defer wg.Done()
			c.consumeTopic(ctx, topic, handler)
		}(topic)
	}
	wg.Wait()

	return nil
}

func (c *natsConsumer) consumeTopic(ctx context.Context, topic string, handler eventbus.EventHandler) {
	stream := streamNameFor(topic)
	durable := durableNameFor(c.groupID)

	consumer, err := c.js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:       durable,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		FilterSubject: topic,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to create nats consumer for topic %s (stream=%s, durable=%s): %v", topic, stream, durable, err)
		return
	}

	iter, err := consumer.Messages()
	if err != nil {
		logger.Errorf(ctx, "failed to start consuming topic %s: %v", topic, err)
		return
	}
	defer iter.Stop()

	go func() {
		<-ctx.Done()
		iter.Stop()
	}()

	for {
		msg, err := iter.Next()
		if err != nil {
			if ctx.Err() != nil {
				logger.Infof(ctx, "consumer context cancelled, stopping subscription to %s", topic)
				return
			}
			logger.Errorf(ctx, "error fetching message from topic %s: %v", topic, err)
			continue
		}

		var event eventbus.Event
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			logger.Errorf(ctx, "failed to unmarshal event from topic %s: %v", topic, err)
			// terminate delivery of the malformed message to avoid blocking the consumer
			if termErr := msg.Term(); termErr != nil {
				logger.Errorf(ctx, "failed to terminate malformed message: %v", termErr)
			}
			continue
		}

		logger.Debugf(ctx, "received event %s (id=%s) from topic %s", event.Type, event.ID, topic)

		if err := handler(ctx, event); err != nil {
			logger.Errorf(ctx, "handler error for event %s (id=%s): %v", event.Type, event.ID, err)
			if nakErr := msg.Nak(); nakErr != nil {
				logger.Errorf(ctx, "failed to nak message for event %s (id=%s): %v", event.Type, event.ID, nakErr)
			}
			continue
		}

		if err := msg.Ack(); err != nil {
			logger.Errorf(ctx, "failed to ack message for event %s (id=%s): %v", event.Type, event.ID, err)
		}
	}
}

func (c *natsConsumer) Close() error {
	logger.Infof(context.TODO(), "closing nats consumer for group '%s'...", c.groupID)
	c.conn.Close()
	return nil
}
