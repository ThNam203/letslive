package push

import (
	"context"
	"errors"
	"sync"
	"time"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

// Defaults applied by NewAsync for zero-valued options.
const (
	// DefaultWorkers is the number of concurrent deliveries.
	DefaultWorkers = 8
	// DefaultQueueSize is how many pushes may wait before new ones are dropped.
	DefaultQueueSize = 256
	// DefaultTimeout bounds a single delivery.
	DefaultTimeout = 10 * time.Second
)

var (
	ErrQueueFull = errors.New("push queue full")
	ErrClosed    = errors.New("push sender closed")
)

type AsyncOptions struct {
	Workers   int
	QueueSize int
	// Timeout bounds each delivery.
	Timeout time.Duration
}

type job struct {
	ctx    context.Context
	userID uuid.UUID
	msg    Message
	ref    string
}

// Async runs deliveries on a fixed worker pool with a bounded queue. Enqueueing
// never blocks: when the queue is full the push is dropped and an error is
// returned (push is best effort, the notification row is already stored).
type Async struct {
	inner   MessageSender
	timeout time.Duration
	jobs    chan job

	mu     sync.RWMutex
	closed bool
	wg     sync.WaitGroup
}

// NewAsync starts the worker pool around inner. Zero option values fall back
// to the Default constants. Call Close to stop the workers.
func NewAsync(inner MessageSender, opts AsyncOptions) *Async {
	if opts.Workers <= 0 {
		opts.Workers = DefaultWorkers
	}
	if opts.QueueSize <= 0 {
		opts.QueueSize = DefaultQueueSize
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	a := &Async{inner: inner, timeout: opts.Timeout, jobs: make(chan job, opts.QueueSize)}
	a.wg.Add(opts.Workers)
	for i := 0; i < opts.Workers; i++ {
		go a.worker()
	}
	return a
}

// Send enqueues a push for a stored notification.
func (a *Async) Send(ctx context.Context, userID uuid.UUID, n domains.Notification) error {
	return a.enqueue(ctx, userID, FromNotification(n), n.Id.String())
}

// SendMessage enqueues a push for an arbitrary message.
func (a *Async) SendMessage(ctx context.Context, userID uuid.UUID, msg Message) error {
	return a.enqueue(ctx, userID, msg, "")
}

func (a *Async) enqueue(ctx context.Context, userID uuid.UUID, msg Message, ref string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.closed {
		return ErrClosed
	}
	// keep log fields (trace ids) but not the request's cancellation
	j := job{ctx: context.WithoutCancel(ctx), userID: userID, msg: msg, ref: ref}
	select {
	case a.jobs <- j:
		return nil
	default:
		return ErrQueueFull
	}
}

// Close stops accepting work and waits for queued deliveries to finish.
func (a *Async) Close() {
	a.mu.Lock()
	if !a.closed {
		a.closed = true
		close(a.jobs)
	}
	a.mu.Unlock()
	a.wg.Wait()
}

func (a *Async) worker() {
	defer a.wg.Done()
	for j := range a.jobs {
		a.run(j)
	}
}

func (a *Async) run(j job) {
	ctx, cancel := context.WithTimeout(j.ctx, a.timeout)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(ctx, "push worker panicked for notification %q user %s: %v", j.ref, j.userID, r)
		}
	}()
	if err := a.inner.SendMessage(ctx, j.userID, j.msg); err != nil {
		logger.Errorf(ctx, "push failed for notification %q user %s: %v", j.ref, j.userID, err)
	}
}
