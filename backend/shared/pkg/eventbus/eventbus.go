package eventbus

import (
	"context"
	"encoding/json"
	"time"
)

// Event represents a domain event with metadata and a typed payload.
// This struct is engine-agnostic — it can be serialized and transported
// over Kafka, NATS, Redis Streams, RabbitMQ, or any other message broker.
type Event struct {
	// ID is a unique identifier for the event (UUID).
	ID string `json:"id"`
	// Type identifies the kind of event (e.g., "stream.started", "user.followed").
	Type string `json:"type"`
	// Source identifies the service that produced the event.
	Source string `json:"source"`
	// Timestamp is when the event was created.
	Timestamp time.Time `json:"timestamp"`
	// Data holds the event-specific payload as raw JSON.
	Data json.RawMessage `json:"data"`
}

// Producer defines the interface for publishing events to a message broker.
// Implementations are engine-specific (Kafka, NATS, Redis Streams, etc.).
type Producer interface {
	// Publish sends an event to the specified topic.
	// The key is used for ordering/partition routing (e.g., userId, streamId).
	// Engines that don't support keyed routing may ignore the key.
	Publish(ctx context.Context, topic string, key string, event Event) error
	// Close gracefully shuts down the producer.
	Close() error
}

// Consumer defines the interface for consuming events from a message broker.
// Implementations are engine-specific (Kafka, NATS, Redis Streams, etc.).
type Consumer interface {
	// Subscribe starts consuming events from the given topics.
	// The handler is called for each received event.
	// This method blocks until the context is cancelled or an unrecoverable error occurs.
	Subscribe(ctx context.Context, topics []string, handler EventHandler) error
	// Close gracefully shuts down the consumer.
	Close() error
}

// EventHandler is a callback function invoked for each consumed event.
// Returning an error signals a processing failure (the message will not be acknowledged).
type EventHandler func(ctx context.Context, event Event) error

// TopicConfig holds the configuration for creating/ensuring a topic exists.
// Engine-specific fields (like partition count) are optional and may be
// ignored by engines that don't support them.
type TopicConfig struct {
	// Name is the topic/subject/channel name.
	Name string
	// NumPartitions is the desired partition count (Kafka-specific, ignored by others).
	NumPartitions int
	// ReplicationFactor is the desired replication factor (Kafka-specific, ignored by others).
	ReplicationFactor int
}

// Admin defines the interface for managing topics on the message broker.
// Not all engines require explicit topic management — implementations may
// no-op if the engine handles topics automatically.
type Admin interface {
	// EnsureTopics creates the specified topics if they do not already exist.
	EnsureTopics(ctx context.Context, topics []TopicConfig) error
	// Close gracefully shuts down the admin connection.
	Close() error
}
