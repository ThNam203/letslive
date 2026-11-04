package push

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/gofrs/uuid/v5"
	"github.com/nats-io/nats.go"
)

// ErrPresenceUnavailable means the presence feed is not connected, so the
// tracker's view cannot be trusted.
var ErrPresenceUnavailable = errors.New("presence feed unavailable")

// PresenceTracker mirrors the realtime service's online/offline events (the
// same NATS subject chat consumes) into an in-memory set. "Online" means the
// user has at least one live realtime socket on any client.
type PresenceTracker struct {
	connected func() bool

	mu     sync.RWMutex
	online map[string]struct{}
}

// NewPresenceTracker takes a function reporting whether the feed is connected.
func NewPresenceTracker(connected func() bool) *PresenceTracker {
	return &PresenceTracker{connected: connected, online: make(map[string]struct{})}
}

// Subscribe starts consuming presence events. Events missed while NATS is
// down cannot be replayed, so the set is cleared on disconnect: unknown users
// count as offline and get the push (a duplicate beats a lost one). This
// installs the connection's disconnect handler.
func (p *PresenceTracker) Subscribe(conn *nats.Conn) (*nats.Subscription, error) {
	conn.SetDisconnectErrHandler(func(_ *nats.Conn, err error) {
		logger.Warnf(context.Background(), "nats disconnected, resetting push presence state: %v", err)
		p.Reset()
	})
	return conn.Subscribe(realtime.PresenceSubject, func(msg *nats.Msg) { p.handle(msg.Data) })
}

func (p *PresenceTracker) handle(data []byte) {
	var event realtime.PresenceEvent
	if err := json.Unmarshal(data, &event); err != nil || event.UserID == "" {
		logger.Warnf(context.Background(), "dropping malformed presence event: %q", data)
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if event.Online {
		p.online[event.UserID] = struct{}{}
	} else {
		delete(p.online, event.UserID)
	}
}

// IsOnline implements PresenceChecker.
func (p *PresenceTracker) IsOnline(_ context.Context, userID uuid.UUID) (bool, error) {
	if !p.connected() {
		return false, ErrPresenceUnavailable
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.online[userID.String()]
	return ok, nil
}

// Reset forgets all presence state.
func (p *PresenceTracker) Reset() {
	p.mu.Lock()
	p.online = make(map[string]struct{})
	p.mu.Unlock()
}

// len returns the number of users currently tracked as online.
func (p *PresenceTracker) len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.online)
}
