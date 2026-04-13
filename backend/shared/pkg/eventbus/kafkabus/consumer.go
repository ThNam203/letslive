package kafkabus

import (
	"context"
	"encoding/json"
	"fmt"
	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type kafkaConsumer struct {
	readers []*kafka.Reader
	brokers []string
	groupID string
}

// NewConsumer creates a new Kafka-backed event consumer for the given consumer group.
func NewConsumer(brokers []string, groupID string) eventbus.Consumer {
	logger.Infof(context.TODO(), "kafka consumer initialized for group '%s' with brokers: %v", groupID, brokers)

	return &kafkaConsumer{
		brokers: brokers,
		groupID: groupID,
	}
}

func (c *kafkaConsumer) Subscribe(ctx context.Context, topics []string, handler eventbus.EventHandler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     c.groupID,
		GroupTopics: topics,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
	})
	c.readers = append(c.readers, reader)

	logger.Infof(ctx, "subscribing to topics %v as group '%s'", topics, c.groupID)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Infof(ctx, "consumer context cancelled, stopping subscription")
				return nil
			}
			logger.Errorf(ctx, "error fetching message: %v", err)
			continue
		}

		var event eventbus.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Errorf(ctx, "failed to unmarshal event from topic %s partition %d offset %d: %v",
				msg.Topic, msg.Partition, msg.Offset, err)
			// commit the malformed message to avoid blocking the consumer
			if commitErr := reader.CommitMessages(ctx, msg); commitErr != nil {
				logger.Errorf(ctx, "failed to commit malformed message: %v", commitErr)
			}
			continue
		}

		logger.Debugf(ctx, "received event %s (id=%s) from topic %s partition %d offset %d",
			event.Type, event.ID, msg.Topic, msg.Partition, msg.Offset)

		if err := handler(ctx, event); err != nil {
			logger.Errorf(ctx, "handler error for event %s (id=%s): %v", event.Type, event.ID, err)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			logger.Errorf(ctx, "failed to commit message for event %s (id=%s): %v", event.Type, event.ID, err)
		}
	}
}

func (c *kafkaConsumer) Close() error {
	logger.Infof(context.TODO(), "closing kafka consumer for group '%s'...", c.groupID)

	var closeErr error
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			closeErr = fmt.Errorf("failed to close reader: %w", err)
			logger.Errorf(context.TODO(), "failed to close kafka reader: %v", err)
		}
	}

	return closeErr
}
