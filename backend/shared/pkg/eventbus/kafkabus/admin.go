package kafkabus

import (
	"context"
	"fmt"
	"net"
	"sen1or/letslive/shared/pkg/eventbus"
	"sen1or/letslive/shared/pkg/logger"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaAdmin struct {
	brokers []string
}

// NewAdmin creates a new Kafka-backed admin for topic management.
func NewAdmin(brokers []string) eventbus.Admin {
	return &kafkaAdmin{
		brokers: brokers,
	}
}

func (a *kafkaAdmin) EnsureTopics(ctx context.Context, topics []eventbus.TopicConfig) error {
	var conn *kafka.Conn
	var err error

	retryDelay := 2 * time.Second
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		conn, err = kafka.DialContext(ctx, "tcp", a.brokers[0])
		if err == nil {
			break
		}

		logger.Warnf(ctx, "failed to connect to kafka broker %s (attempt %d/%d): %v - retrying in %v...",
			a.brokers[0], i+1, maxRetries, err, retryDelay)

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("context cancelled while connecting to kafka: %w", ctx.Err())
		case <-timer.C:
		}

		retryDelay *= 2
		if retryDelay > 30*time.Second {
			retryDelay = 30 * time.Second
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to kafka after %d attempts: %w", maxRetries, err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get kafka controller: %w", err)
	}

	controllerConn, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to kafka controller: %w", err)
	}
	defer controllerConn.Close()

	var topicConfigs []kafka.TopicConfig
	for _, t := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             t.Name,
			NumPartitions:     t.NumPartitions,
			ReplicationFactor: t.ReplicationFactor,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	for _, t := range topics {
		logger.Infof(ctx, "ensured kafka topic '%s' (partitions=%d, replication=%d)", t.Name, t.NumPartitions, t.ReplicationFactor)
	}

	return nil
}

func (a *kafkaAdmin) Close() error {
	return nil
}
