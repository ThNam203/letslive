package hub

import (
	"context"
	"fmt"
	"sync"

	"sen1or/letslive/realtime/protocol"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"
)

type Subscription interface {
	Unsubscribe() error
}

type Source interface {
	Subscribe(subject string, handler func(data []byte)) (Subscription, error)
}

type Subscriber interface {
	UserID() string
	Deliver(frame []byte)
}

type topicState struct {
	subscription Subscription
	subscribers  map[Subscriber]struct{}
	members      map[string]int
}

type Hub struct {
	mu        sync.Mutex
	source    Source
	publisher realtime.Publisher
	topics    map[realtime.Topic]*topicState
}

func New(source Source, publisher realtime.Publisher) *Hub {
	return &Hub{
		source:    source,
		publisher: publisher,
		topics:    make(map[realtime.Topic]*topicState),
	}
}

func (h *Hub) Subscribe(ctx context.Context, s Subscriber, topic realtime.Topic) error {
	h.mu.Lock()

	state, exists := h.topics[topic]
	if !exists {
		subscription, err := h.source.Subscribe(topic.Subject(), func(data []byte) { h.dispatch(topic, data) })
		if err != nil {
			h.mu.Unlock()
			return fmt.Errorf("subscribe %s: %w", topic.String(), err)
		}
		state = &topicState{
			subscription: subscription,
			subscribers:  make(map[Subscriber]struct{}),
			members:      make(map[string]int),
		}
		h.topics[topic] = state
	}

	if _, already := state.subscribers[s]; already {
		h.mu.Unlock()
		return nil
	}
	state.subscribers[s] = struct{}{}
	joined := trackMember(state, topic, s.UserID(), 1)
	h.mu.Unlock()

	if joined {
		h.publishMember(ctx, topic, realtime.EventMemberJoined, s.UserID())
	}
	return nil
}

func (h *Hub) Unsubscribe(ctx context.Context, s Subscriber, topic realtime.Topic) {
	h.mu.Lock()

	state, exists := h.topics[topic]
	if !exists {
		h.mu.Unlock()
		return
	}
	if _, subscribed := state.subscribers[s]; !subscribed {
		h.mu.Unlock()
		return
	}

	delete(state.subscribers, s)
	left := trackMember(state, topic, s.UserID(), -1)

	// unsubscribing under the lock keeps an in-flight message from the old
	// subscription from reaching subscribers of a re-created topic
	if len(state.subscribers) == 0 {
		delete(h.topics, topic)
		if err := state.subscription.Unsubscribe(); err != nil {
			logger.Errorf(ctx, "failed to unsubscribe %s: %v", topic.String(), err)
		}
	}
	h.mu.Unlock()

	if left {
		h.publishMember(ctx, topic, realtime.EventMemberLeft, s.UserID())
	}
}

func (h *Hub) dispatch(topic realtime.Topic, data []byte) {
	ctx := context.Background()

	msg, err := realtime.DecodeMessage(data)
	if err != nil {
		logger.Warnf(ctx, "dropping malformed realtime message on %s: %v", topic.String(), err)
		return
	}
	frame, err := protocol.EventFrame(topic, msg)
	if err != nil {
		logger.Errorf(ctx, "failed to build event frame for %s: %v", topic.String(), err)
		return
	}

	h.mu.Lock()
	state, exists := h.topics[topic]
	var targets []Subscriber
	if exists {
		targets = make([]Subscriber, 0, len(state.subscribers))
		for s := range state.subscribers {
			targets = append(targets, s)
		}
	}
	h.mu.Unlock()

	for _, s := range targets {
		s.Deliver(frame)
	}
}

func (h *Hub) publishMember(ctx context.Context, topic realtime.Topic, eventType string, userID string) {
	if err := h.publisher.Publish(ctx, topic, eventType, realtime.MemberEvent{UserID: userID}); err != nil {
		logger.Errorf(ctx, "failed to publish %s on %s: %v", eventType, topic.String(), err)
	}
}

// trackMember counts authenticated sockets per user in a room and reports
// whether the user just joined (0→1) or left (1→0).
func trackMember(state *topicState, topic realtime.Topic, userID string, delta int) bool {
	if topic.Kind != realtime.TopicKindRoom || userID == "" {
		return false
	}

	state.members[userID] += delta
	count := state.members[userID]
	if count <= 0 {
		delete(state.members, userID)
		return delta < 0
	}
	return delta > 0 && count == 1
}
