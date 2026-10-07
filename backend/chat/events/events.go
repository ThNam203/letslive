// Package events defines the DM pushes the chat service sends through the
// realtime gateway. Event names and payloads are the ones the Node /dm-ws
// socket sent, minus the type and recipientIds fields that the gateway
// envelope now carries.
package events

import (
	"context"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"
)

const (
	DmNewMessage        = "dm:new_message"
	DmUserTyping        = "dm:user_typing"
	DmUserStoppedTyping = "dm:user_stopped_typing"
	DmReadReceipt       = "dm:read_receipt"
	DmUserOnline        = "dm:user_online"
	DmUserOffline       = "dm:user_offline"
)

type NewMessage struct {
	ConversationID string                 `json:"conversationId"`
	Message        *dto.DmMessageResponse `json:"message"`
}

type Typing struct {
	ConversationID string `json:"conversationId"`
	UserID         string `json:"userId"`
	Username       string `json:"username"`
}

type ReadReceipt struct {
	ConversationID string  `json:"conversationId"`
	UserID         string  `json:"userId"`
	MessageID      *string `json:"messageId,omitempty"`
	ReadAt         string  `json:"readAt"`
}

type Presence struct {
	UserID string `json:"userId"`
}

// ChatMessage is a live chat line as the web's ReceivedMessage expects it;
// timestamp is epoch milliseconds, as the Node socket sent it.
type ChatMessage struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userId"`
	Username       string  `json:"username"`
	Text           string  `json:"text"`
	ProfilePicture *string `json:"profilePicture"`
	Timestamp      int64   `json:"timestamp"`
}

// Notifier pushes an event to each user's topic. Pushes are hints: the data is
// already saved, so a failure is logged and never fails the caller.
type Notifier struct {
	publisher realtime.Publisher
}

func NewNotifier(publisher realtime.Publisher) *Notifier {
	return &Notifier{publisher: publisher}
}

func (n *Notifier) ToUsers(ctx context.Context, userIDs []string, eventType string, data any) {
	for _, userID := range userIDs {
		if err := n.publisher.Publish(ctx, realtime.UserTopic(userID), eventType, data); err != nil {
			logger.Errorf(ctx, "failed to push %s to user %s: %v", eventType, userID, err)
		}
	}
}

func (n *Notifier) ToRoom(ctx context.Context, roomID string, eventType string, data any) {
	if err := n.publisher.Publish(ctx, realtime.RoomTopic(roomID), eventType, data); err != nil {
		logger.Errorf(ctx, "failed to push %s to room %s: %v", eventType, roomID, err)
	}
}
