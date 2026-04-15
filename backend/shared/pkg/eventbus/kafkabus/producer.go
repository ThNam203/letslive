package kafkabus

import (
	"context"
	"encoding/json"
	"fmt"
	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka-backed event producer.
func NewProducer(brokers []string) eventbus.Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  3,
		Async:        false,
	}

	logger.Infof(context.TODO(), "kafka producer initialized with brokers: %v", brokers)

	return &kafkaProducer{
		writer: w,
	}
}

func (p *kafkaProducer) Publish(ctx context.Context, topic string, key string, event eventbus.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.Type)},
			{Key: "event-source", Value: []byte(event.Source)},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		logger.Errorf(ctx, "failed to publish event %s to topic %s: %v", event.Type, topic, err)
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}

	logger.Debugf(ctx, "published event %s (id=%s) to topic %s with key %s", event.Type, event.ID, topic, key)
	return nil
}

func (p *kafkaProducer) Close() error {
	logger.Infof(context.TODO(), "closing kafka producer...")
	return p.writer.Close()
}
