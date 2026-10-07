package presence

import (
	"context"
	"encoding/json"
	"time"

	"sen1or/letslive/chat/events"
	"sen1or/letslive/chat/repositories"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/nats-io/nats.go"
)

const (
	fanOutTimeout = 5 * time.Second
	// the Node DM socket looked up contacts in the 100 most recently
	// updated conversations
	contactConversationLimit = 100
)

// Subscriber turns the gateway's online/offline signal into dm:user_online /
// dm:user_offline pushes to everyone the user has a conversation with. Only
// the chat service knows who those contacts are.
type Subscriber struct {
	conversations *repositories.ConversationRepository
	notifier      *events.Notifier
}

func NewSubscriber(conversations *repositories.ConversationRepository, notifier *events.Notifier) *Subscriber {
	return &Subscriber{conversations: conversations, notifier: notifier}
}

// Subscribe handles events inline in the NATS callback, which keeps one
// user's online/offline events in order.
func (s *Subscriber) Subscribe(conn *nats.Conn) (*nats.Subscription, error) {
	return conn.Subscribe(realtime.PresenceSubject, s.handle)
}

func (s *Subscriber) handle(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), fanOutTimeout)
	defer cancel()

	var event realtime.PresenceEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil || event.UserID == "" {
		logger.Warnf(ctx, "dropping malformed presence event: %q", msg.Data)
		return
	}

	conversations, err := s.conversations.ListForUser(ctx, event.UserID, 0, contactConversationLimit)
	if err != nil {
		logger.Errorf(ctx, "failed to load contacts of %s for presence: %v", event.UserID, err)
		return
	}

	seen := map[string]struct{}{}
	contacts := []string{}
	for i := range conversations {
		for _, id := range conversations[i].OtherParticipantIDs(event.UserID) {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				contacts = append(contacts, id)
			}
		}
	}

	eventType := events.DmUserOffline
	if event.Online {
		eventType = events.DmUserOnline
	}
	s.notifier.ToUsers(ctx, contacts, eventType, events.Presence{UserID: event.UserID})
}
