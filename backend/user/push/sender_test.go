package push

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

type fakeTokenStore struct {
	mu      sync.Mutex
	tokens  map[uuid.UUID][]domains.DeviceToken
	deleted []string
	listErr error
}

func (f *fakeTokenStore) ListTokensByUserID(_ context.Context, userID uuid.UUID) ([]domains.DeviceToken, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tokens[userID], f.listErr
}

func (f *fakeTokenStore) DeleteByToken(_ context.Context, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, token)
	return nil
}

func (f *fakeTokenStore) deletedTokens() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleted...)
}

type fakeGateway struct {
	mu       sync.Mutex
	calls    [][]string
	messages []Message
	status   map[string]TokenStatus
	err      error
}

func (g *fakeGateway) SendToTokens(_ context.Context, tokens []string, msg Message) ([]TokenResult, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls = append(g.calls, append([]string(nil), tokens...))
	g.messages = append(g.messages, msg)
	if g.err != nil {
		return nil, g.err
	}
	results := make([]TokenResult, len(tokens))
	for i, tok := range tokens {
		st, ok := g.status[tok]
		if !ok {
			st = StatusSent
		}
		results[i] = TokenResult{Token: tok, Status: st}
	}
	return results, nil
}

func (g *fakeGateway) callCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.calls)
}

type fakePresence struct {
	online bool
	err    error
}

func (p fakePresence) IsOnline(context.Context, uuid.UUID) (bool, error) { return p.online, p.err }

func newID() uuid.UUID { return uuid.Must(uuid.NewV4()) }

func tokenRow(userID uuid.UUID, token string) domains.DeviceToken {
	return domains.DeviceToken{UserId: userID, Token: token, Platform: "ios"}
}

func sampleNotification(userID uuid.UUID) domains.Notification {
	ref := newID()
	url := "/gifts/1"
	return domains.Notification{
		Id: newID(), UserId: userID, Type: "gift_received",
		Title: "Gift", Message: "You got a gift", ReferenceId: &ref, ActionUrl: &url,
	}
}

func TestSendsToEveryTokenOfTheUser(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{
		user: {tokenRow(user, "a"), tokenRow(user, "b")},
	}}
	gw := &fakeGateway{}
	svc := NewService(store, fakePresence{}, gw)

	if err := svc.Send(context.Background(), user, sampleNotification(user)); err != nil {
		t.Fatal(err)
	}

	if gw.callCount() != 1 || len(gw.calls[0]) != 2 {
		t.Fatalf("calls = %v, want one call with both tokens", gw.calls)
	}
}

func TestSkipsSendWhenUserIsOnlineOverRealtime(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{user: {tokenRow(user, "a")}}}
	gw := &fakeGateway{}
	svc := NewService(store, fakePresence{online: true}, gw)

	if err := svc.Send(context.Background(), user, sampleNotification(user)); err != nil {
		t.Fatal(err)
	}

	if gw.callCount() != 0 {
		t.Fatalf("gateway called %d times for an online user", gw.callCount())
	}
}

func TestSendsWhenPresenceIsUnavailable(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{user: {tokenRow(user, "a")}}}
	gw := &fakeGateway{}
	svc := NewService(store, fakePresence{err: errors.New("nats down")}, gw)

	if err := svc.Send(context.Background(), user, sampleNotification(user)); err != nil {
		t.Fatal(err)
	}

	if gw.callCount() != 1 {
		t.Fatalf("gateway calls = %d, want 1 (a duplicate beats a lost push)", gw.callCount())
	}
}

func TestDeletesTokenWhenFCMReportsUnregistered(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{
		user: {tokenRow(user, "good"), tokenRow(user, "stale")},
	}}
	gw := &fakeGateway{status: map[string]TokenStatus{"stale": StatusUnregistered}}
	svc := NewService(store, fakePresence{}, gw)

	if err := svc.Send(context.Background(), user, sampleNotification(user)); err != nil {
		t.Fatal(err)
	}

	got := store.deletedTokens()
	if len(got) != 1 || got[0] != "stale" {
		t.Fatalf("deleted = %v, want [stale]", got)
	}
}

func TestInvalidArgumentTokenDeletedOnlyWhenAnotherTokenSucceeded(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{
		user: {tokenRow(user, "good"), tokenRow(user, "bad")},
	}}
	gw := &fakeGateway{status: map[string]TokenStatus{"bad": StatusInvalidArgument}}
	if err := NewService(store, fakePresence{}, gw).Send(context.Background(), user, sampleNotification(user)); err != nil {
		t.Fatal(err)
	}
	if got := store.deletedTokens(); len(got) != 1 || got[0] != "bad" {
		t.Fatalf("deleted = %v, want [bad]", got)
	}

	// every token rejected: the message itself may be at fault, keep tokens
	store2 := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{user: {tokenRow(user, "only")}}}
	gw2 := &fakeGateway{status: map[string]TokenStatus{"only": StatusInvalidArgument}}
	_ = NewService(store2, fakePresence{}, gw2).Send(context.Background(), user, sampleNotification(user))
	if got := store2.deletedTokens(); len(got) != 0 {
		t.Fatalf("deleted = %v, want none", got)
	}
}

func TestNoTokensMeansNoCall(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{}}
	gw := &fakeGateway{}
	svc := NewService(store, fakePresence{}, gw)

	if err := svc.Send(context.Background(), user, sampleNotification(user)); err != nil {
		t.Fatal(err)
	}

	if gw.callCount() != 0 {
		t.Fatalf("gateway called %d times with no tokens", gw.callCount())
	}
}

func TestGatewayFailureIsReturnedAndKeepsTokens(t *testing.T) {
	user := newID()
	store := &fakeTokenStore{tokens: map[uuid.UUID][]domains.DeviceToken{user: {tokenRow(user, "a")}}}
	gw := &fakeGateway{err: errors.New("fcm 503")}
	svc := NewService(store, fakePresence{}, gw)

	if err := svc.Send(context.Background(), user, sampleNotification(user)); err == nil {
		t.Fatal("expected the gateway error")
	}
	if len(store.deletedTokens()) != 0 {
		t.Fatal("tokens must survive a transport failure")
	}
}

func TestFromNotificationCarriesPayloadFields(t *testing.T) {
	user := newID()
	n := sampleNotification(user)

	msg := FromNotification(n)

	if msg.Title != n.Title || msg.Body != n.Message {
		t.Fatalf("title/body = %q/%q", msg.Title, msg.Body)
	}
	want := map[string]string{"type": "gift_received", "referenceId": n.ReferenceId.String(), "actionUrl": "/gifts/1"}
	for k, v := range want {
		if msg.Data[k] != v {
			t.Errorf("data[%s] = %q, want %q", k, msg.Data[k], v)
		}
	}
}

func TestFromNotificationOmitsEmptyOptionalFields(t *testing.T) {
	msg := FromNotification(domains.Notification{Type: "system", Title: "t", Message: "m"})

	if _, ok := msg.Data["referenceId"]; ok {
		t.Error("referenceId must be absent when nil")
	}
	if _, ok := msg.Data["actionUrl"]; ok {
		t.Error("actionUrl must be absent when nil")
	}
	if msg.Data["type"] != "system" {
		t.Errorf("type = %q", msg.Data["type"])
	}
}

func TestDisabledSenderIsANoOp(t *testing.T) {
	sender := NewDisabled()
	if err := sender.Send(context.Background(), newID(), sampleNotification(newID())); err != nil {
		t.Fatalf("disabled Send returned %v", err)
	}
	if err := sender.SendMessage(context.Background(), newID(), Message{Title: "x"}); err != nil {
		t.Fatalf("disabled SendMessage returned %v", err)
	}
}

type blockingSender struct {
	started  chan struct{}
	release  chan struct{}
	inFlight atomic.Int32
	maxSeen  atomic.Int32
	done     atomic.Int32
}

func (b *blockingSender) SendMessage(_ context.Context, _ uuid.UUID, _ Message) error {
	cur := b.inFlight.Add(1)
	for {
		prev := b.maxSeen.Load()
		if cur <= prev || b.maxSeen.CompareAndSwap(prev, cur) {
			break
		}
	}
	b.started <- struct{}{}
	<-b.release
	b.inFlight.Add(-1)
	b.done.Add(1)
	return nil
}

func TestAsyncBoundsConcurrencyAndDropsWhenQueueFull(t *testing.T) {
	inner := &blockingSender{started: make(chan struct{}, 100), release: make(chan struct{})}
	async := NewAsync(inner, AsyncOptions{Workers: 2, QueueSize: 3, Timeout: time.Second})

	var queued, dropped int
	for i := 0; i < 20; i++ {
		if err := async.SendMessage(context.Background(), newID(), Message{}); err != nil {
			dropped++
		} else {
			queued++
		}
	}
	<-inner.started
	<-inner.started
	if got := inner.maxSeen.Load(); got > 2 {
		t.Fatalf("max concurrent sends = %d, want <= 2", got)
	}
	if dropped == 0 {
		t.Fatal("expected overflow to be rejected instead of blocking the caller")
	}

	close(inner.release)
	async.Close()
	if int(inner.done.Load()) != queued {
		t.Fatalf("completed %d of %d accepted jobs", inner.done.Load(), queued)
	}
}

func TestAsyncSendDoesNotBlockTheCaller(t *testing.T) {
	inner := &blockingSender{started: make(chan struct{}, 10), release: make(chan struct{})}
	async := NewAsync(inner, AsyncOptions{Workers: 1, QueueSize: 2, Timeout: time.Second})

	returned := make(chan struct{})
	go func() {
		_ = async.SendMessage(context.Background(), newID(), Message{})
		close(returned)
	}()
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("SendMessage blocked on a slow sender")
	}
	close(inner.release)
	async.Close()
}

type ctxState struct {
	err         error
	hasDeadline bool
}

type ctxCapturingSender struct {
	got     chan ctxState
	proceed chan struct{}
}

func (c ctxCapturingSender) SendMessage(ctx context.Context, _ uuid.UUID, _ Message) error {
	<-c.proceed // the request context is cancelled before this inspection
	_, ok := ctx.Deadline()
	c.got <- ctxState{err: ctx.Err(), hasDeadline: ok}
	return nil
}

func TestAsyncDetachesFromRequestCancellationAndBoundsTime(t *testing.T) {
	inner := ctxCapturingSender{got: make(chan ctxState, 1), proceed: make(chan struct{})}
	async := NewAsync(inner, AsyncOptions{Workers: 1, QueueSize: 1, Timeout: 5 * time.Second})
	reqCtx, cancel := context.WithCancel(context.Background())

	if err := async.SendMessage(reqCtx, newID(), Message{}); err != nil {
		t.Fatal(err)
	}
	cancel()
	close(inner.proceed)
	state := <-inner.got
	async.Close()

	if errors.Is(state.err, context.Canceled) {
		t.Fatal("job context inherited the request cancellation")
	}
	if !state.hasDeadline {
		t.Fatal("job context has no timeout")
	}
}

type funcSender func(context.Context, uuid.UUID, Message) error

func (f funcSender) SendMessage(ctx context.Context, u uuid.UUID, m Message) error {
	return f(ctx, u, m)
}

func TestAsyncSurvivesPanickingSenderAndKeepsWorking(t *testing.T) {
	var ok atomic.Int32
	inner := funcSender(func(_ context.Context, _ uuid.UUID, m Message) error {
		if m.Title == "panic" {
			panic("boom")
		}
		ok.Add(1)
		return nil
	})
	async := NewAsync(inner, AsyncOptions{Workers: 1, QueueSize: 4, Timeout: time.Second})

	_ = async.SendMessage(context.Background(), newID(), Message{Title: "panic"})
	_ = async.SendMessage(context.Background(), newID(), Message{Title: "fine"})
	async.Close()

	if ok.Load() != 1 {
		t.Fatalf("worker did not survive the panic: ok = %d", ok.Load())
	}
}

func TestAsyncRejectsAfterClose(t *testing.T) {
	async := NewAsync(funcSender(func(context.Context, uuid.UUID, Message) error { return nil }),
		AsyncOptions{Workers: 1, QueueSize: 1, Timeout: time.Second})
	async.Close()

	if err := async.SendMessage(context.Background(), newID(), Message{}); err == nil {
		t.Fatal("expected an error after Close")
	}
}
