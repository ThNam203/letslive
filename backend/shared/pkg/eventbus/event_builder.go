package eventbus

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
)

// NewEvent creates a new Event with a generated ID and current timestamp.
// The data parameter will be marshaled to JSON.
func NewEvent(eventType string, source string, data any) (Event, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return Event{}, fmt.Errorf("failed to generate event id: %w", err)
	}

	rawData, err := json.Marshal(data)
	if err != nil {
		return Event{}, fmt.Errorf("failed to marshal event data: %w", err)
	}

	return Event{
		ID:        id.String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Data:      rawData,
	}, nil
}

// ParseEventData unmarshals the raw JSON data of an event into the target type.
func ParseEventData[T any](event Event) (*T, error) {
	var target T
	if err := json.Unmarshal(event.Data, &target); err != nil {
		return nil, fmt.Errorf("failed to parse event data for event type '%s': %w", event.Type, err)
	}
	return &target, nil
}
