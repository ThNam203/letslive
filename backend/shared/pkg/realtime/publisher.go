package realtime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(ctx context.Context, topic Topic, eventType string, data any) error
}

type PresencePublisher interface {
	PublishPresence(ctx context.Context, event PresenceEvent) error
}

type NATSPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(conn *nats.Conn) *NATSPublisher {
	return &NATSPublisher{conn: conn}
}

func (p *NATSPublisher) Publish(_ context.Context, topic Topic, eventType string, data any) error {
	if !topic.Valid() {
		return fmt.Errorf("publish to %q: %w", topic.String(), ErrInvalidTopic)
	}
	body, err := EncodeMessage(eventType, data)
	if err != nil {
		return err
	}
	return p.conn.Publish(topic.Subject(), body)
}

func (p *NATSPublisher) PublishPresence(_ context.Context, event PresenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode presence event: %w", err)
	}
	return p.conn.Publish(PresenceSubject, body)
}
