package realtime

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	EventNotificationCreated = "notification.created"
	EventMemberJoined        = "member.joined"
	EventMemberLeft          = "member.left"
)

var ErrEmptyEventType = errors.New("realtime message has no type")

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MemberEvent struct {
	UserID string `json:"userId"`
}

type PresenceEvent struct {
	UserID string `json:"userId"`
	Online bool   `json:"online"`
}

func EncodeMessage(eventType string, data any) ([]byte, error) {
	if eventType == "" {
		return nil, ErrEmptyEventType
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encode %s data: %w", eventType, err)
	}
	return json.Marshal(Message{Type: eventType, Data: raw})
}

func DecodeMessage(body []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return Message{}, fmt.Errorf("decode realtime message: %w", err)
	}
	if msg.Type == "" {
		return Message{}, ErrEmptyEventType
	}
	return msg, nil
}
