package protocol

import (
	"encoding/json"
	"errors"

	"sen1or/letslive/shared/pkg/realtime"
)

const (
	OpSubscribe   = "subscribe"
	OpUnsubscribe = "unsubscribe"
	OpPing        = "ping"
	OpEvent       = "event"
	OpPong        = "pong"
	OpError       = "error"
)

const (
	ErrCodeInvalidTopic         = "invalid_topic"
	ErrCodeForbiddenTopic       = "forbidden_topic"
	ErrCodeTooManySubscriptions = "too_many_subscriptions"
	ErrCodeBadFrame             = "bad_frame"
	ErrCodeUnavailable          = "unavailable"
)

var ErrBadFrame = errors.New("bad frame")

type ClientFrame struct {
	Op    string `json:"op"`
	Topic string `json:"topic,omitempty"`
}

type ServerFrame struct {
	Op    string          `json:"op"`
	Topic string          `json:"topic,omitempty"`
	Type  string          `json:"type,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
	Code  string          `json:"code,omitempty"`
}

func ParseClientFrame(raw []byte) (ClientFrame, error) {
	var frame ClientFrame
	if err := json.Unmarshal(raw, &frame); err != nil {
		return ClientFrame{}, ErrBadFrame
	}

	switch frame.Op {
	case OpPing:
		return frame, nil
	case OpSubscribe, OpUnsubscribe:
		if frame.Topic == "" {
			return ClientFrame{}, ErrBadFrame
		}
		return frame, nil
	default:
		return ClientFrame{}, ErrBadFrame
	}
}

func EventFrame(topic realtime.Topic, msg realtime.Message) ([]byte, error) {
	return json.Marshal(ServerFrame{Op: OpEvent, Topic: topic.String(), Type: msg.Type, Data: msg.Data})
}

func ErrorFrame(code string) []byte {
	return staticFrame(ServerFrame{Op: OpError, Code: code})
}

func PongFrame() []byte {
	return staticFrame(ServerFrame{Op: OpPong})
}

// frames built only from constant strings cannot fail to marshal
func staticFrame(frame ServerFrame) []byte {
	body, err := json.Marshal(frame)
	if err != nil {
		panic(err)
	}
	return body
}
