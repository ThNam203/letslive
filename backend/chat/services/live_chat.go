package services

import (
	"context"
	"slices"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/events"
	"sen1or/letslive/chat/gateway/userservice"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/repositories"
	"sen1or/letslive/shared/pkg/realtime"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// LiveChatService stores live chat lines and pushes them to the room topic,
// which replaces the Node /ws socket and its Redis fan-out.
type LiveChatService struct {
	messages *repositories.LiveMessageRepository
	users    *userservice.Gateway
	notifier *events.Notifier
}

func NewLiveChatService(messages *repositories.LiveMessageRepository, users *userservice.Gateway, notifier *events.Notifier) *LiveChatService {
	return &LiveChatService{messages: messages, users: users, notifier: notifier}
}

// History returns the room's latest messages, oldest first.
func (s *LiveChatService) History(ctx context.Context, roomID string) ([]domains.LiveMessage, error) {
	messages, err := s.messages.Latest(ctx, roomID, domains.LiveHistorySize)
	if err != nil {
		return nil, err
	}
	slices.Reverse(messages)
	return messages, nil
}

// Send saves a line from the signed-in user and pushes chat.message to
// everyone watching the room. The sender's name and avatar come from the
// user service, never from the request.
func (s *LiveChatService) Send(ctx context.Context, userID, roomID, text string) (*domains.LiveMessage, error) {
	if !realtime.RoomTopic(roomID).Valid() {
		return nil, domains.ErrRoomNotFound
	}
	if len(jsutil.Trim(text)) == 0 || jsutil.Length(text) > domains.MaxLiveMessageLength {
		return nil, domains.ErrInvalidInput
	}

	identities, err := s.users.GetIdentities(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	sender, ok := identities[userID]
	if !ok {
		return nil, domains.ErrInvalidInput
	}
	if !hasFinishedSetup(sender) {
		return nil, domains.ErrUserSetupIncomplete
	}

	message := &domains.LiveMessage{
		ID:             bson.NewObjectID(),
		RoomID:         roomID,
		Username:       sender.Username,
		UserID:         userID,
		Text:           text,
		ProfilePicture: sender.ProfilePicture,
		Timestamp:      domains.Now(),
		Version:        0,
	}
	if err := s.messages.Insert(ctx, message); err != nil {
		return nil, err
	}

	s.notifier.ToRoom(ctx, roomID, realtime.EventChatMessage, events.ChatMessage{
		ID:             message.ID.Hex(),
		UserID:         message.UserID,
		Username:       message.Username,
		Text:           message.Text,
		ProfilePicture: message.ProfilePicture,
		Timestamp:      message.Timestamp.UnixMilli(),
	})
	return message, nil
}
