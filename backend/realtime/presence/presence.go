package presence

import (
	"context"
	"sync"
	"time"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"
)

type pendingOffline struct {
	timer      *time.Timer
	generation uint64
}

type Tracker struct {
	mu         sync.Mutex
	publisher  realtime.PresencePublisher
	grace      time.Duration
	counts     map[string]int
	pending    map[string]pendingOffline
	generation uint64
}

func NewTracker(publisher realtime.PresencePublisher, grace time.Duration) *Tracker {
	return &Tracker{
		publisher: publisher,
		grace:     grace,
		counts:    make(map[string]int),
		pending:   make(map[string]pendingOffline),
	}
}

func (t *Tracker) Connected(ctx context.Context, userID string) {
	t.mu.Lock()
	t.counts[userID]++
	first := t.counts[userID] == 1
	pending, wasPending := t.pending[userID]
	if first && wasPending {
		pending.timer.Stop()
		delete(t.pending, userID)
	}
	t.mu.Unlock()

	if first && !wasPending {
		t.publish(ctx, userID, true)
	}
}

func (t *Tracker) Disconnected(_ context.Context, userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.counts[userID] == 0 {
		return
	}
	t.counts[userID]--
	if t.counts[userID] > 0 {
		return
	}
	delete(t.counts, userID)

	t.generation++
	generation := t.generation
	t.pending[userID] = pendingOffline{
		generation: generation,
		timer:      time.AfterFunc(t.grace, func() { t.expire(userID, generation) }),
	}
}

// expire ignores timers that a reconnect already cancelled but that fired
// before Stop could prevent it, by comparing generations.
func (t *Tracker) expire(userID string, generation uint64) {
	t.mu.Lock()
	pending, exists := t.pending[userID]
	current := exists && pending.generation == generation
	if current {
		delete(t.pending, userID)
	}
	t.mu.Unlock()

	if current {
		t.publish(context.Background(), userID, false)
	}
}

func (t *Tracker) publish(ctx context.Context, userID string, online bool) {
	event := realtime.PresenceEvent{UserID: userID, Online: online}
	if err := t.publisher.PublishPresence(ctx, event); err != nil {
		logger.Errorf(ctx, "failed to publish presence for %s: %v", userID, err)
	}
}
