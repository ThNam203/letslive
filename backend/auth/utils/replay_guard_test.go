package utils

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sen1or/letslive/shared/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Error)
	os.Exit(m.Run())
}

func TestMemoryReplayGuard_secondUseIsRejected(t *testing.T) {
	g := NewMemoryReplayGuard(10)
	exp := time.Now().Add(time.Hour)
	if !g.Consume("a", exp) {
		t.Fatal("first use must pass")
	}
	if g.Consume("a", exp) {
		t.Fatal("second use must be rejected")
	}
}

func TestMemoryReplayGuard_differentTokensBothPass(t *testing.T) {
	g := NewMemoryReplayGuard(10)
	exp := time.Now().Add(time.Hour)
	if !g.Consume("a", exp) || !g.Consume("b", exp) {
		t.Fatal("distinct keys must both pass")
	}
}

func TestMemoryReplayGuard_expiredEntriesAreSwept(t *testing.T) {
	now := time.Now()
	g := newMemoryReplayGuard(10, func() time.Time { return now })
	g.Consume("a", now.Add(time.Minute))
	g.Consume("b", now.Add(time.Hour))

	now = now.Add(2 * time.Minute) // past a's exp and past the sweep interval
	if !g.Consume("c", now.Add(time.Hour)) {
		t.Fatal("new key must pass")
	}
	if got := g.size(); got != 2 { // a swept; b and c remain
		t.Fatalf("size = %d, want 2", got)
	}
	if g.Consume("b", now.Add(time.Hour)) {
		t.Fatal("unexpired key must still be rejected after a sweep")
	}
}

func TestMemoryReplayGuard_boundedAndFailsClosed(t *testing.T) {
	g := NewMemoryReplayGuard(2)
	exp := time.Now().Add(time.Hour)
	if !g.Consume("a", exp) || !g.Consume("b", exp) {
		t.Fatal("entries under the cap must pass")
	}
	if g.Consume("c", exp) {
		t.Fatal("at the cap with nothing expired the guard must fail closed")
	}
}

func TestMemoryReplayGuard_concurrentUseLetsExactlyOneThrough(t *testing.T) {
	g := NewMemoryReplayGuard(10)
	exp := time.Now().Add(time.Hour)
	var passed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if g.Consume("same", exp) {
				passed.Add(1)
			}
		}()
	}
	wg.Wait()
	if passed.Load() != 1 {
		t.Fatalf("%d goroutines passed, want exactly 1", passed.Load())
	}
}
