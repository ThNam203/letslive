package utils

import (
	"sync"
	"time"
)

// DefaultReplayGuardMaxEntries bounds the in-memory guard.
const DefaultReplayGuardMaxEntries = 100_000

// ReplayGuard makes a verified App Check token single-use. A standard App
// Check token is reusable until it expires, so without this one token taken
// from a real device could be replayed for credential stuffing or OTP mail
// spam. Mobile clients should send limited-use App Check tokens.
type ReplayGuard interface {
	// Consume records key until expiresAt and reports whether this is its
	// first use. It must be safe for concurrent use and must fail closed
	// (return false) when it cannot record the key.
	Consume(key string, expiresAt time.Time) bool
}

// memoryReplayGuard keeps consumed keys in process memory.
//
// It is per instance: with several auth instances behind a load balancer a
// token can be replayed once per instance, so a multi-instance deployment
// needs a shared store (for example Redis) behind the same ReplayGuard
// interface. Entries expire at the token's exp and are swept lazily; at the
// size cap it fails closed rather than grow.
type memoryReplayGuard struct {
	mu         sync.Mutex
	seen       map[string]time.Time
	maxEntries int
	now        func() time.Time
	nextSweep  time.Time
}

const replaySweepInterval = time.Minute

func NewMemoryReplayGuard(maxEntries int) ReplayGuard {
	return newMemoryReplayGuard(maxEntries, time.Now)
}

func newMemoryReplayGuard(maxEntries int, now func() time.Time) *memoryReplayGuard {
	return &memoryReplayGuard{seen: make(map[string]time.Time), maxEntries: maxEntries, now: now}
}

func (g *memoryReplayGuard) Consume(key string, expiresAt time.Time) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.now()
	if now.After(g.nextSweep) || len(g.seen) >= g.maxEntries {
		g.sweepLocked(now)
	}

	if exp, used := g.seen[key]; used && exp.After(now) {
		return false
	}
	if len(g.seen) >= g.maxEntries {
		return false
	}
	g.seen[key] = expiresAt
	return true
}

func (g *memoryReplayGuard) sweepLocked(now time.Time) {
	for k, exp := range g.seen {
		if !exp.After(now) {
			delete(g.seen, k)
		}
	}
	g.nextSweep = now.Add(replaySweepInterval)
}

func (g *memoryReplayGuard) size() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.seen)
}
