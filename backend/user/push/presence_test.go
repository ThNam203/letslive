package push

import (
	"context"
	"testing"
)

func presenceEvent(userID string, online bool) []byte {
	if online {
		return []byte(`{"userId":"` + userID + `","online":true}`)
	}
	return []byte(`{"userId":"` + userID + `","online":false}`)
}

func TestPresenceTracksOnlineAndOffline(t *testing.T) {
	p := NewPresenceTracker(func() bool { return true })
	user := newID()

	p.handle(presenceEvent(user.String(), true))
	if on, err := p.IsOnline(context.Background(), user); err != nil || !on {
		t.Fatalf("IsOnline = %v, %v; want true", on, err)
	}

	p.handle(presenceEvent(user.String(), false))
	if on, _ := p.IsOnline(context.Background(), user); on {
		t.Fatal("user still online after offline event")
	}
}

func TestPresenceIgnoresMalformedEvents(t *testing.T) {
	p := NewPresenceTracker(func() bool { return true })
	p.handle([]byte(`not json`))
	p.handle([]byte(`{"online":true}`))
	if p.count() != 0 {
		t.Fatalf("tracked %d users from malformed events", p.count())
	}
}

func TestPresenceReportsErrorAndOfflineWhenDisconnected(t *testing.T) {
	connected := true
	p := NewPresenceTracker(func() bool { return connected })
	user := newID()
	p.handle(presenceEvent(user.String(), true))

	connected = false
	on, err := p.IsOnline(context.Background(), user)

	if on || err == nil {
		t.Fatalf("IsOnline = %v, %v; want false with an error while NATS is down", on, err)
	}
}

func TestPresenceResetDropsStaleState(t *testing.T) {
	p := NewPresenceTracker(func() bool { return true })
	user := newID()
	p.handle(presenceEvent(user.String(), true))

	p.Reset()

	if on, _ := p.IsOnline(context.Background(), user); on {
		t.Fatal("state survived Reset")
	}
}
