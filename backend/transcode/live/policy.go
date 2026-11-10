package live

import (
	"context"
	"sync"
)

// Backpressure decides what a full internal queue does with the next item.
type Backpressure interface {
	// giveUp is closed once a sender should stop waiting for room.
	giveUp() <-chan struct{}
}

// dropWhenFull drops at once: a livestream must never wait on ffmpeg or the
// recorder.
type dropWhenFull struct{}

func (dropWhenFull) giveUp() <-chan struct{} { return alreadyClosed }

var alreadyClosed = func() chan struct{} {
	c := make(chan struct{})
	close(c)
	return c
}()

// waitUntil waits for room until ctx ends: a file must lose nothing, and it
// can be read as slowly as the encoder needs.
type waitUntil struct{ ctx context.Context }

func (w waitUntil) giveUp() <-chan struct{} { return w.ctx.Done() }

// send queues v on ch and reports whether it was queued, waiting for room as
// bp allows.
func send[T any](bp Backpressure, ch chan<- T, v T) bool {
	select {
	case ch <- v:
		return true
	default:
	}
	select {
	case ch <- v:
		return true
	case <-bp.giveUp():
		return false
	}
}

// Failures is told about everything that makes a recording incomplete: a
// dropped quality, a lost segment, unusable input.
type Failures interface {
	Report(err error)
	// Err is the error that fails the session, or nil.
	Err() error
}

// logOnly keeps a livestream going: the call sites have already logged the
// problem and degraded (dropped a rendition, skipped a packet).
type logOnly struct{}

func (logOnly) Report(error) {}
func (logOnly) Err() error   { return nil }

// firstError fails the session on the first report and cancels its work, so
// an upload is retried instead of being marked ready with something missing.
type firstError struct {
	cancel context.CancelFunc

	mu  sync.Mutex
	err error
}

func (f *firstError) Report(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return
	}
	f.err = err
	f.cancel()
}

func (f *firstError) Err() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.err
}
