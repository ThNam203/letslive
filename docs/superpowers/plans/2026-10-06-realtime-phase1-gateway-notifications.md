# Realtime Phase 1: Gateway + Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the Go realtime gateway and push `notification.created` to the web, replacing the 30s notification polling.

**Architecture:** A new Go service `backend/realtime` accepts one WebSocket per browser tab at `/realtime`, verifies the `ACCESS_TOKEN` cookie itself, auto-subscribes authenticated sockets to `user:<id>`, and fans out messages that services publish on core NATS subjects `rt.user.<id>` / `rt.room.<id>`. The user service publishes after it saves a notification. The web app holds the socket in a `RealtimeProvider` and invalidates notification queries on push and on reconnect.

**Tech Stack:** Go 1.26, `github.com/coder/websocket` v1.8.15, `github.com/nats-io/nats.go` v1.52.0 (core NATS, no JetStream), `github.com/golang-jwt/jwt/v5` v5.3.1, Next.js 16 + React 19 + TanStack Query 5, Kong, Consul, Spring config server (`letslive-configs` repo).

**Spec:** `docs/superpowers/specs/2026-10-06-realtime-gateway-design.md`. This plan covers spec phase 1 only. Phases 2–4 get their own plans after this one lands.

## Global Constraints

- Gateway port `7785`; Kong path `/realtime`; Consul/service name `realtime`; config profile name `realtime_service`.
- NATS subjects: `rt.user.<id>`, `rt.room.<id>`, `rt.presence`. Client topics: `user:<id>`, `room:<id>`. Topic ids must match `^[A-Za-z0-9-]{1,36}$`.
- Core NATS only. Do not use `shared/pkg/eventbus` / JetStream for realtime pushes.
- Socket limits: inbound frame 1 KiB, 50 subscriptions per socket (`user:` auto topic excluded), 64-frame outbound buffer, 10s write timeout, 30s server ping, 5s presence offline grace.
- Wire frames: client `{"op":"subscribe|unsubscribe|ping","topic":...}`; server `{"op":"event","topic","type","data"}`, `{"op":"pong"}`, `{"op":"error","code"}`. Error codes: `invalid_topic`, `forbidden_topic`, `too_many_subscriptions`, `bad_frame`, `unavailable`.
- Allowed origins (config `websocket.allowedOrigins`): dev `localhost:3000`, `localhost:5000`; prod `letslive.work`.
- **No unit/integration tests in this work** (owner writes them later). Verify with `go build`, `go vet`, `npx tsc --noEmit`, `npm run lint` and the running-stack checks in Task 10.
- **No commits until the owner reviews the running app.** Each task ends with a checkpoint, not a commit.
- Comments only for technical concerns (why a non-obvious mechanism exists); no narration or business-rule comments.
- TypeScript: no `any`. Go: handle every error (log it when it cannot be returned).
- New dependencies allowed: `github.com/coder/websocket` only. Nothing else.

## Review Focus

These failure modes are implied by the spec but not exercised by any automated check. Each one has a manual check in the task that owns the code.

1. **A browser on an allowed origin behind Kong** (`Origin: http://localhost:3000`, request Host = gateway). Must upgrade. **A foreign origin** must get 403. Checked in Task 10 step 3.
2. **An expired or forged `ACCESS_TOKEN` cookie** must still connect as anonymous, without the `user:` topic. It must not be rejected and must not be treated as a logged-in user. Checked in Task 10 step 4.
3. **Two tabs of the same user.** Both receive the notification, and closing one tab keeps the other subscribed. Checked in Task 10 step 6.
4. **A non-JSON or typeless message on `rt.user.<id>`** from a buggy publisher. It must be dropped with a warning while the gateway keeps serving. Checked in Task 10 step 7.
5. **NATS restart while sockets are open.** `nats.go` re-subscribes, and pushes resume without a gateway restart. Checked in Task 10 step 8.

---

## File Map

| File | Status | Responsibility |
|---|---|---|
| `backend/shared/pkg/natsconn/conn.go` | create | Core NATS dial with retry/backoff |
| `backend/shared/pkg/eventbus/natsbus/conn.go` | modify | Reuse `natsconn.Connect`, keep JetStream setup |
| `backend/shared/pkg/realtime/topic.go` | create | `Topic`, validation, subject mapping |
| `backend/shared/pkg/realtime/message.go` | create | Event type names, NATS message encode/decode, payload structs |
| `backend/shared/pkg/realtime/publisher.go` | create | `Publisher` / `PresencePublisher` interfaces and the NATS implementation |
| `backend/user/config/config.go` | modify | `nats.url` config |
| `backend/user/cmd/main.go` | modify | Connect NATS, inject publisher, drain on shutdown |
| `backend/user/services/notification.go` | modify | Publish `notification.created` after create |
| `backend/realtime/go.mod`, `backend/go.work` | create/modify | New module |
| `backend/realtime/config/config.go` | create | Config struct, `PostProcess` |
| `backend/realtime/protocol/protocol.go` | create | Frame types, parse/build |
| `backend/realtime/auth/auth.go` | create | Verified cookie → user id |
| `backend/realtime/hub/hub.go` | create | Topic registry, ref-counted source subscriptions, fan-out, membership events |
| `backend/realtime/hub/natssource.go` | create | `Source` over core NATS |
| `backend/realtime/presence/presence.go` | create | Per-user socket count, offline grace |
| `backend/realtime/client/client.go` | create | One socket: read loop, write loop, cleanup |
| `backend/realtime/api/server.go` | create | HTTP server, health, upgrade |
| `backend/realtime/cmd/main.go` | create | Bootstrap |
| `backend/realtime/Dockerfile` | create | Image build |
| `docker-compose-dev.yaml`, `docker-compose.yaml` | modify | `realtime` service; `user` depends on `nats` |
| `configs/kong.yml` | modify | `Realtime` service + `/realtime` route |
| `../letslive-configs/realtime_service-{dev,prod}.yml` | create | Gateway config |
| `../letslive-configs/user_service-{dev,prod}.yml` | modify | `nats.url` |
| `web/global.ts` | modify | One `getWsUrl(path)` builder, `REALTIME_URL` |
| `web/constant/realtime.ts` | create | Event names, reconnect delays |
| `web/types/realtime.ts` | create | Server frame types |
| `web/lib/realtime/parse-frame.ts` | create | Untrusted JSON → typed frame |
| `web/contexts/realtime-context.tsx` | create | `RealtimeProvider`, `useRealtime` |
| `web/hooks/use-notification-realtime.ts` | create | Invalidate notification queries on push/reconnect |
| `web/hooks/queries/use-notifications.ts` | modify | Drop polling, export root key |
| `web/app/(main)/_components/header/notification-bell.tsx` | modify | Use the realtime hook |
| `web/app/(main)/layout.tsx` | modify | Mount `RealtimeProvider` |

---

### Task 0: Branch

- [ ] **Step 1: Check the working tree**

Run: `cd /Users/ictsaigon.vn/mywork/letslive && git status --short`

The current branch `fix/settings-chat-gift-bugs` has unrelated uncommitted changes (`docs/ISSUES.md`, `web/app/opengraph-image.tsx`, `web/lib/i18n/index.ts`, `web/AGENTS.md`, `web/CLAUDE.md`), and the spec and this plan are untracked. **Ask the owner** what to do with the unrelated changes before switching branches. Do not stash, commit or discard them yourself.

- [ ] **Step 2: Create the branch from fresh `main`**

```bash
git fetch origin main
git switch -c feat/realtime-gateway origin/main
```

Untracked files (spec, plan) travel with the switch.

---

### Task 1: Shared NATS connect + realtime contract

**Files:**
- Create: `backend/shared/pkg/natsconn/conn.go`
- Modify: `backend/shared/pkg/eventbus/natsbus/conn.go` (whole file)
- Create: `backend/shared/pkg/realtime/topic.go`, `message.go`, `publisher.go`

**Interfaces:**
- Produces:
  - `natsconn.Connect(ctx context.Context, url string) (*nats.Conn, error)`
  - `realtime.Topic{Kind, ID string}`
  - `realtime.UserTopic(id) Topic`, `realtime.RoomTopic(id) Topic`
  - `(Topic).String()`, `(Topic).Subject()`, `(Topic).Valid()`
  - `realtime.ParseTopic(string) (Topic, error)`, `realtime.ErrInvalidTopic`
  - `realtime.TopicKindUser`, `realtime.TopicKindRoom`, `realtime.PresenceSubject`
  - `realtime.Message{Type string; Data json.RawMessage}`
  - `realtime.EncodeMessage(eventType string, data any) ([]byte, error)`, `realtime.DecodeMessage([]byte) (Message, error)`
  - `realtime.MemberEvent{UserID}`, `realtime.PresenceEvent{UserID, Online}`
  - constants `EventNotificationCreated`, `EventMemberJoined`, `EventMemberLeft`
  - `realtime.Publisher` interface `Publish(ctx, Topic, eventType string, data any) error`
  - `realtime.PresencePublisher` interface `PublishPresence(ctx, PresenceEvent) error`
  - `realtime.NewNATSPublisher(*nats.Conn) *NATSPublisher` (implements both)

- [ ] **Step 1: Create `backend/shared/pkg/natsconn/conn.go`**

```go
package natsconn

import (
	"context"
	"fmt"
	"time"

	"sen1or/letslive/shared/pkg/logger"

	"github.com/nats-io/nats.go"
)

const (
	maxConnectAttempts = 10
	initialRetryDelay  = 2 * time.Second
	maxRetryDelay      = 30 * time.Second
	reconnectWait      = 2 * time.Second
)

// Connect dials NATS with retry because the server is often not ready yet when
// a service starts alongside it in docker-compose. Once connected, nats.go
// reconnects forever and restores subscriptions on its own.
func Connect(ctx context.Context, url string) (*nats.Conn, error) {
	retryDelay := initialRetryDelay
	var lastErr error

	for attempt := 1; attempt <= maxConnectAttempts; attempt++ {
		conn, err := nats.Connect(url, nats.MaxReconnects(-1), nats.ReconnectWait(reconnectWait))
		if err == nil {
			return conn, nil
		}
		lastErr = err

		logger.Warnf(ctx, "failed to connect to nats at %s (attempt %d/%d): %v - retrying in %v...",
			url, attempt, maxConnectAttempts, err, retryDelay)

		timer := time.NewTimer(retryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("context cancelled while connecting to nats: %w", ctx.Err())
		case <-timer.C:
		}

		retryDelay = min(retryDelay*2, maxRetryDelay)
	}

	return nil, fmt.Errorf("failed to connect to nats after %d attempts: %w", maxConnectAttempts, lastErr)
}
```

- [ ] **Step 2: Replace `backend/shared/pkg/eventbus/natsbus/conn.go`**

```go
package natsbus

import (
	"context"
	"fmt"

	"sen1or/letslive/shared/pkg/natsconn"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func connect(ctx context.Context, url string) (*nats.Conn, jetstream.JetStream, error) {
	conn, err := natsconn.Connect(ctx, url)
	if err != nil {
		return nil, nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	return conn, js, nil
}
```

Retry behaviour is unchanged: the same 2s initial delay, doubling to a 30s cap, 10 attempts, and `MaxReconnects(-1)` / `ReconnectWait(2s)`.

- [ ] **Step 3: Create `backend/shared/pkg/realtime/topic.go`**

```go
package realtime

import (
	"errors"
	"regexp"
	"strings"
)

const (
	TopicKindUser = "user"
	TopicKindRoom = "room"

	PresenceSubject = "rt.presence"

	subjectPrefix = "rt"
)

var (
	ErrInvalidTopic = errors.New("invalid topic")

	// ids become NATS subject tokens; anything outside this set could inject
	// wildcards ("*", ">") or extra tokens (".")
	topicIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,36}$`)
)

type Topic struct {
	Kind string
	ID   string
}

func UserTopic(userID string) Topic { return Topic{Kind: TopicKindUser, ID: userID} }

func RoomTopic(roomID string) Topic { return Topic{Kind: TopicKindRoom, ID: roomID} }

func (t Topic) String() string { return t.Kind + ":" + t.ID }

func (t Topic) Subject() string { return subjectPrefix + "." + t.Kind + "." + t.ID }

func (t Topic) Valid() bool {
	return (t.Kind == TopicKindUser || t.Kind == TopicKindRoom) && topicIDPattern.MatchString(t.ID)
}

func ParseTopic(raw string) (Topic, error) {
	kind, id, found := strings.Cut(raw, ":")
	topic := Topic{Kind: kind, ID: id}
	if !found || !topic.Valid() {
		return Topic{}, ErrInvalidTopic
	}
	return topic, nil
}
```

- [ ] **Step 4: Create `backend/shared/pkg/realtime/message.go`**

```go
package realtime

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	EventNotificationCreated = "notification.created"
	EventMemberJoined        = "member.joined"
	EventMemberLeft          = "member.left"
)

var ErrEmptyEventType = errors.New("realtime message has no type")

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MemberEvent struct {
	UserID string `json:"userId"`
}

type PresenceEvent struct {
	UserID string `json:"userId"`
	Online bool   `json:"online"`
}

func EncodeMessage(eventType string, data any) ([]byte, error) {
	if eventType == "" {
		return nil, ErrEmptyEventType
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encode %s data: %w", eventType, err)
	}
	return json.Marshal(Message{Type: eventType, Data: raw})
}

func DecodeMessage(body []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return Message{}, fmt.Errorf("decode realtime message: %w", err)
	}
	if msg.Type == "" {
		return Message{}, ErrEmptyEventType
	}
	return msg, nil
}
```

- [ ] **Step 5: Create `backend/shared/pkg/realtime/publisher.go`**

```go
package realtime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(ctx context.Context, topic Topic, eventType string, data any) error
}

type PresencePublisher interface {
	PublishPresence(ctx context.Context, event PresenceEvent) error
}

type NATSPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(conn *nats.Conn) *NATSPublisher {
	return &NATSPublisher{conn: conn}
}

func (p *NATSPublisher) Publish(_ context.Context, topic Topic, eventType string, data any) error {
	if !topic.Valid() {
		return fmt.Errorf("publish to %q: %w", topic.String(), ErrInvalidTopic)
	}
	body, err := EncodeMessage(eventType, data)
	if err != nil {
		return err
	}
	return p.conn.Publish(topic.Subject(), body)
}

func (p *NATSPublisher) PublishPresence(_ context.Context, event PresenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode presence event: %w", err)
	}
	return p.conn.Publish(PresenceSubject, body)
}
```

- [ ] **Step 6: Verify**

Run: `cd backend/shared && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 7: Checkpoint.** Leave uncommitted.

---

### Task 2: User service publishes `notification.created`

**Files:**
- Modify: `backend/user/config/config.go`
- Modify: `backend/user/cmd/main.go`
- Modify: `backend/user/services/notification.go`

**Interfaces:**
- Consumes: `natsconn.Connect`, `realtime.NewNATSPublisher`, `realtime.Publisher`, `realtime.UserTopic`, `realtime.EventNotificationCreated` (Task 1)
- Produces: `services.NewNotificationService(repo domains.NotificationRepository, publisher realtime.Publisher) *NotificationService`. The NATS message is `{"type":"notification.created","data":<domains.Notification JSON>}` on `rt.user.<recipientId>`. The `data` shape is the same JSON as one item of `GET /v1/user/me/notifications`.

- [ ] **Step 1: Add NATS config in `backend/user/config/config.go`**

Add the type after `Tracer`:

```go
type NATS struct {
	URL string `yaml:"url"`
}
```

Change `Config` to:

```go
type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	MinIO    `yaml:"minio"`
	Tracer   `yaml:"tracer"`
	NATS     `yaml:"nats"`
}
```

At the end of `PostProcess`, before `return nil`:

```go
	if config.NATS.URL == "" {
		return fmt.Errorf("nats.url is not configured")
	}
```

- [ ] **Step 2: Publish in `backend/user/services/notification.go`**

Replace the struct, the constructor and the last line of `CreateNotification`:

```go
type NotificationService struct {
	notificationRepo domains.NotificationRepository
	publisher        realtime.Publisher
}

func NewNotificationService(
	notificationRepo domains.NotificationRepository,
	publisher realtime.Publisher,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		publisher:        publisher,
	}
}
```

In `CreateNotification`, replace `return s.notificationRepo.Create(ctx, notification)` with:

```go
	created, err := s.notificationRepo.Create(ctx, notification)
	if err != nil {
		return nil, err
	}

	// the row is already stored and the client refetches on reconnect, so a
	// failed push must not fail the request
	topic := realtime.UserTopic(created.UserId.String())
	if pubErr := s.publisher.Publish(ctx, topic, realtime.EventNotificationCreated, created); pubErr != nil {
		logger.Errorf(ctx, "failed to publish notification %s to %s: %v", created.Id, topic.String(), pubErr)
	}

	return created, nil
```

Add imports: `"sen1or/letslive/shared/pkg/logger"` and `"sen1or/letslive/shared/pkg/realtime"`.

Parity: validation, the repo call and the returned values are unchanged. The only addition is the publish after a successful create. Both callers (`GiftService.notifyRecipient` and `CreateNotificationInternalHandler`) go through this method, so both now push.

- [ ] **Step 3: Wire it in `backend/user/cmd/main.go`**

After `dbConn := sharedutils.ConnectDB(...)` / `defer dbConn.Close()`:

```go
	natsConn, err := natsconn.Connect(ctx, config.NATS.URL)
	if err != nil {
		logger.Panicf(ctx, "failed to connect to nats: %v", err)
	}
	realtimePublisher := realtime.NewNATSPublisher(natsConn)

	server := SetupServer(ctx, dbConn, registry, config, realtimePublisher)
```

(This replaces the existing `server := SetupServer(ctx, dbConn, registry, config)` line.)

Inside the shutdown section, add a fourth goroutine next to the others:

```go
	shutdownWg.Add(1)
	go (func() {
		if err := natsConn.Drain(); err != nil {
			logger.Errorf(shutdownCtx, "failed to drain nats connection: %v", err)
		}
		shutdownWg.Done()
	})()
```

Change `SetupServer`'s signature and the notification service line:

```go
func SetupServer(ctx context.Context, dbConn *pgxpool.Pool, registry discovery.Registry, cfg *cfg.Config, publisher realtime.Publisher) *api.APIServer {
	...
	var notificationService = services.NewNotificationService(notificationRepo, publisher)
```

Add imports: `"sen1or/letslive/shared/pkg/natsconn"` and `"sen1or/letslive/shared/pkg/realtime"`.

- [ ] **Step 4: Verify**

Run: `cd backend/user && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 5: Checkpoint.** Leave uncommitted.

---

### Task 3: Realtime module, config, protocol, auth

**Files:**
- Create: `backend/realtime/go.mod` (via commands)
- Modify: `backend/go.work`
- Create: `backend/realtime/config/config.go`, `protocol/protocol.go`, `auth/auth.go`

**Interfaces:**
- Consumes: `realtime.Topic`, `realtime.Message`, `realtime.UserTopic` (Task 1)
- Produces:
  - `config.Config` with `Service`, `NATS{URL}`, `WebSocket{AllowedOrigins []string}`, `Tracer`, `AccessTokenSecret string`; `config.PostProcess(*Config) error`
  - `protocol.ParseClientFrame([]byte) (ClientFrame, error)`
  - `protocol.EventFrame(realtime.Topic, realtime.Message) ([]byte, error)`, `protocol.ErrorFrame(code string) []byte`, `protocol.PongFrame() []byte`
  - `protocol.Op*` constants and `protocol.ErrCode*` constants
  - `auth.NewVerifier(secret string) *Verifier`, `(*Verifier).UserID(*http.Request) (string, bool)`

- [ ] **Step 1: Create the module**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/backend
mkdir -p realtime && cd realtime
go mod init sen1or/letslive/realtime
go mod edit -go=1.26.0 -require=sen1or/letslive/shared@v0.0.0
cd .. && go work use ./realtime
cd realtime && go get github.com/coder/websocket@v1.8.15 github.com/golang-jwt/jwt/v5@v5.3.1 github.com/nats-io/nats.go@v1.52.0
```

Expected: `backend/go.work` now lists `./realtime`, and `go.mod` lists the three requires plus `sen1or/letslive/shared v0.0.0`. The workspace `replace` resolves `shared`, the same way it does for `user`. Don't add a `replace` to `go.mod`.

- [ ] **Step 2: Create `backend/realtime/config/config.go`**

```go
package config

import (
	"errors"
	"os"
)

type Service struct {
	Name           string `yaml:"name"`
	Hostname       string `yaml:"hostname"`
	APIBindAddress string `yaml:"apiBindAddress"`
	APIPort        int    `yaml:"apiPort"`
}

type NATS struct {
	URL string `yaml:"url"`
}

type WebSocket struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` // milliseconds
}

type Config struct {
	Service   `yaml:"service"`
	NATS      `yaml:"nats"`
	WebSocket `yaml:"websocket"`
	Tracer    `yaml:"tracer"`

	AccessTokenSecret string `yaml:"-"`
}

func (c Config) GetServiceName() string     { return c.Service.Name }
func (c Config) GetTracerEndpoint() string  { return c.Tracer.Endpoint }
func (c Config) GetTracerBatchTimeout() int { return c.Tracer.BatchTimeout }
func (c Config) IsSecure() bool             { return c.Tracer.Secure }

func PostProcess(config *Config) error {
	secret := os.Getenv("ACCESS_TOKEN_SECRET")
	if secret == "" {
		return errors.New("ACCESS_TOKEN_SECRET is not set")
	}
	config.AccessTokenSecret = secret

	if config.NATS.URL == "" {
		return errors.New("nats.url is not configured")
	}
	if len(config.WebSocket.AllowedOrigins) == 0 {
		return errors.New("websocket.allowedOrigins is not configured")
	}
	return nil
}
```

- [ ] **Step 3: Create `backend/realtime/protocol/protocol.go`**

```go
package protocol

import (
	"encoding/json"
	"errors"

	"sen1or/letslive/shared/pkg/realtime"
)

const (
	OpSubscribe   = "subscribe"
	OpUnsubscribe = "unsubscribe"
	OpPing        = "ping"
	OpEvent       = "event"
	OpPong        = "pong"
	OpError       = "error"
)

const (
	ErrCodeInvalidTopic         = "invalid_topic"
	ErrCodeForbiddenTopic       = "forbidden_topic"
	ErrCodeTooManySubscriptions = "too_many_subscriptions"
	ErrCodeBadFrame             = "bad_frame"
	ErrCodeUnavailable          = "unavailable"
)

var ErrBadFrame = errors.New("bad frame")

type ClientFrame struct {
	Op    string `json:"op"`
	Topic string `json:"topic,omitempty"`
}

type ServerFrame struct {
	Op    string          `json:"op"`
	Topic string          `json:"topic,omitempty"`
	Type  string          `json:"type,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
	Code  string          `json:"code,omitempty"`
}

func ParseClientFrame(raw []byte) (ClientFrame, error) {
	var frame ClientFrame
	if err := json.Unmarshal(raw, &frame); err != nil {
		return ClientFrame{}, ErrBadFrame
	}

	switch frame.Op {
	case OpPing:
		return frame, nil
	case OpSubscribe, OpUnsubscribe:
		if frame.Topic == "" {
			return ClientFrame{}, ErrBadFrame
		}
		return frame, nil
	default:
		return ClientFrame{}, ErrBadFrame
	}
}

func EventFrame(topic realtime.Topic, msg realtime.Message) ([]byte, error) {
	return json.Marshal(ServerFrame{Op: OpEvent, Topic: topic.String(), Type: msg.Type, Data: msg.Data})
}

func ErrorFrame(code string) []byte {
	return staticFrame(ServerFrame{Op: OpError, Code: code})
}

func PongFrame() []byte {
	return staticFrame(ServerFrame{Op: OpPong})
}

// frames built only from constant strings cannot fail to marshal
func staticFrame(frame ServerFrame) []byte {
	body, err := json.Marshal(frame)
	if err != nil {
		panic(err)
	}
	return body
}
```

- [ ] **Step 4: Create `backend/realtime/auth/auth.go`**

```go
package auth

import (
	"net/http"

	"sen1or/letslive/shared/pkg/realtime"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenCookie = "ACCESS_TOKEN"

type accessClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// Verifier checks the signature itself because the /realtime Kong route has
// no JWT plugin (anonymous viewers must be able to connect).
type Verifier struct {
	secret []byte
}

func NewVerifier(secret string) *Verifier {
	return &Verifier{secret: []byte(secret)}
}

// UserID returns the user id of a valid ACCESS_TOKEN cookie; ok is false for
// anonymous, expired or forged tokens.
func (v *Verifier) UserID(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(accessTokenCookie)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	var claims accessClaims
	_, err = jwt.ParseWithClaims(
		cookie.Value,
		&claims,
		func(*jwt.Token) (any, error) { return v.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", false
	}

	if !realtime.UserTopic(claims.UserID).Valid() {
		return "", false
	}
	return claims.UserID, true
}
```

Parity: the claims shape (`userId` + registered claims, HS256, `ACCESS_TOKEN_SECRET`) matches `auth/services/jwt.go` `generateAccessToken`.

- [ ] **Step 5: Verify**

Run: `cd backend/realtime && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 6: Checkpoint.** Leave uncommitted.

---

### Task 4: Hub and NATS source

**Files:**
- Create: `backend/realtime/hub/hub.go`, `backend/realtime/hub/natssource.go`

**Interfaces:**
- Consumes: `realtime.Topic`, `realtime.Publisher`, `realtime.DecodeMessage`, `realtime.MemberEvent`, `realtime.EventMember*` (Task 1); `protocol.EventFrame` (Task 3)
- Produces:
  - `hub.Subscriber` interface `{ UserID() string; Deliver(frame []byte) }`
  - `hub.Source` interface `{ Subscribe(subject string, handler func([]byte)) (Subscription, error) }`, `hub.Subscription` interface `{ Unsubscribe() error }`
  - `hub.New(Source, realtime.Publisher) *Hub`
  - `(*Hub).Subscribe(ctx, Subscriber, realtime.Topic) error`, `(*Hub).Unsubscribe(ctx, Subscriber, realtime.Topic)`
  - `hub.NewNATSSource(*nats.Conn) *NATSSource`

- [ ] **Step 1: Create `backend/realtime/hub/hub.go`**

```go
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
```

- [ ] **Step 2: Create `backend/realtime/hub/natssource.go`**

```go
package hub

import "github.com/nats-io/nats.go"

type NATSSource struct {
	conn *nats.Conn
}

func NewNATSSource(conn *nats.Conn) *NATSSource {
	return &NATSSource{conn: conn}
}

func (s *NATSSource) Subscribe(subject string, handler func(data []byte)) (Subscription, error) {
	subscription, err := s.conn.Subscribe(subject, func(msg *nats.Msg) { handler(msg.Data) })
	if err != nil {
		return nil, err
	}
	return subscription, nil
}
```

- [ ] **Step 3: Verify**

Run: `cd backend/realtime && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 4: Checkpoint.** Leave uncommitted.

---

### Task 5: Presence tracker

**Files:**
- Create: `backend/realtime/presence/presence.go`

**Interfaces:**
- Consumes: `realtime.PresencePublisher`, `realtime.PresenceEvent` (Task 1)
- Produces: `presence.NewTracker(realtime.PresencePublisher, grace time.Duration) *Tracker`, `(*Tracker).Connected(ctx, userID string)`, `(*Tracker).Disconnected(ctx, userID string)`

- [ ] **Step 1: Create `backend/realtime/presence/presence.go`**

```go
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
```

- [ ] **Step 2: Verify**

Run: `cd backend/realtime && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 3: Checkpoint.** Leave uncommitted.

---

### Task 6: Socket client and HTTP server

**Files:**
- Create: `backend/realtime/client/client.go`, `backend/realtime/api/server.go`

**Interfaces:**
- Consumes: `hub.Subscriber` (Task 4); `protocol.*` (Task 3); `auth.Verifier` (Task 3); `config.Config` (Task 3); `realtime.ParseTopic`, `realtime.UserTopic` (Task 1)
- Produces:
  - `client.Hub` interface (the `Subscribe`/`Unsubscribe` signatures of `*hub.Hub`), `client.Presence` interface (the `Connected`/`Disconnected` signatures of `*presence.Tracker`)
  - `client.Serve(ctx, *websocket.Conn, userID string, Hub, Presence)`
  - `api.NewServer(cfg *config.Config, verifier *auth.Verifier, h client.Hub, p client.Presence, isHealthy func() bool) *Server`, `(*Server).ListenAndServe() error`, `(*Server).Shutdown(ctx) error`

- [ ] **Step 1: Create `backend/realtime/client/client.go`**

```go
package client

import (
	"context"
	"sync"
	"time"

	"sen1or/letslive/realtime/hub"
	"sen1or/letslive/realtime/protocol"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/coder/websocket"
)

const (
	MaxFrameBytes    = 1024
	MaxSubscriptions = 50
	SendBuffer       = 64
	WriteTimeout     = 10 * time.Second
	PingInterval     = 30 * time.Second
)

type Hub interface {
	Subscribe(ctx context.Context, s hub.Subscriber, topic realtime.Topic) error
	Unsubscribe(ctx context.Context, s hub.Subscriber, topic realtime.Topic)
}

type Presence interface {
	Connected(ctx context.Context, userID string)
	Disconnected(ctx context.Context, userID string)
}

type Client struct {
	userID   string
	send     chan []byte
	overflow func()
	// owned by the read loop goroutine; cleanup runs on the same goroutine
	topics map[realtime.Topic]struct{}
}

func (c *Client) UserID() string { return c.userID }

// Deliver never blocks the hub: a socket that cannot keep up is dropped and
// the browser reconnects and refetches.
func (c *Client) Deliver(frame []byte) {
	select {
	case c.send <- frame:
	default:
		c.overflow()
	}
}

func Serve(ctx context.Context, conn *websocket.Conn, userID string, h Hub, p Presence) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var overflowOnce sync.Once
	c := &Client{
		userID:   userID,
		send:     make(chan []byte, SendBuffer),
		overflow: func() { overflowOnce.Do(cancel) },
		topics:   make(map[realtime.Topic]struct{}),
	}
	cleanupCtx := context.WithoutCancel(ctx)

	if userID != "" {
		p.Connected(ctx, userID)
		defer p.Disconnected(cleanupCtx, userID)

		userTopic := realtime.UserTopic(userID)
		if err := h.Subscribe(ctx, c, userTopic); err != nil {
			logger.Errorf(ctx, "failed to subscribe %s: %v", userTopic.String(), err)
			if closeErr := conn.Close(websocket.StatusTryAgainLater, "realtime unavailable"); closeErr != nil {
				logger.Debugf(ctx, "close after failed subscribe: %v", closeErr)
			}
			return
		}
		defer h.Unsubscribe(cleanupCtx, c, userTopic)
	}
	defer c.unsubscribeAll(cleanupCtx, h)

	conn.SetReadLimit(MaxFrameBytes)
	go c.writeLoop(ctx, cancel, conn)
	c.readLoop(ctx, conn, h)

	if err := conn.CloseNow(); err != nil {
		logger.Debugf(ctx, "close socket: %v", err)
	}
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn, h Hub) {
	for {
		messageType, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if messageType != websocket.MessageText {
			c.Deliver(protocol.ErrorFrame(protocol.ErrCodeBadFrame))
			continue
		}
		c.handleFrame(ctx, data, h)
	}
}

func (c *Client) handleFrame(ctx context.Context, data []byte, h Hub) {
	frame, err := protocol.ParseClientFrame(data)
	if err != nil {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeBadFrame))
		return
	}

	switch frame.Op {
	case protocol.OpPing:
		c.Deliver(protocol.PongFrame())
	case protocol.OpSubscribe:
		c.subscribe(ctx, frame.Topic, h)
	case protocol.OpUnsubscribe:
		c.unsubscribe(ctx, frame.Topic, h)
	}
}

func (c *Client) subscribe(ctx context.Context, rawTopic string, h Hub) {
	topic, err := realtime.ParseTopic(rawTopic)
	if err != nil {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeInvalidTopic))
		return
	}
	if topic.Kind != realtime.TopicKindRoom {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeForbiddenTopic))
		return
	}
	if _, already := c.topics[topic]; already {
		return
	}
	if len(c.topics) >= MaxSubscriptions {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeTooManySubscriptions))
		return
	}
	if err := h.Subscribe(ctx, c, topic); err != nil {
		logger.Errorf(ctx, "failed to subscribe %s: %v", topic.String(), err)
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeUnavailable))
		return
	}
	c.topics[topic] = struct{}{}
}

func (c *Client) unsubscribe(ctx context.Context, rawTopic string, h Hub) {
	topic, err := realtime.ParseTopic(rawTopic)
	if err != nil {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeInvalidTopic))
		return
	}
	if _, subscribed := c.topics[topic]; !subscribed {
		return
	}
	h.Unsubscribe(ctx, c, topic)
	delete(c.topics, topic)
}

func (c *Client) unsubscribeAll(ctx context.Context, h Hub) {
	for topic := range c.topics {
		h.Unsubscribe(ctx, c, topic)
	}
}

func (c *Client) writeLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	defer cancel()

	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-c.send:
			writeCtx, done := context.WithTimeout(ctx, WriteTimeout)
			err := conn.Write(writeCtx, websocket.MessageText, frame)
			done()
			if err != nil {
				return
			}
		case <-ticker.C:
			// Kong drops upstream connections that are idle for 60s
			pingCtx, done := context.WithTimeout(ctx, WriteTimeout)
			err := conn.Ping(pingCtx)
			done()
			if err != nil {
				return
			}
		}
	}
}
```

- [ ] **Step 2: Create `backend/realtime/api/server.go`**

```go
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/realtime/auth"
	"sen1or/letslive/realtime/client"
	"sen1or/letslive/realtime/config"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/coder/websocket"
)

const readHeaderTimeout = 10 * time.Second

type Server struct {
	cfg        *config.Config
	verifier   *auth.Verifier
	hub        client.Hub
	presence   client.Presence
	isHealthy  func() bool
	baseCtx    context.Context
	cancelBase context.CancelFunc
	httpServer *http.Server
}

func NewServer(cfg *config.Config, verifier *auth.Verifier, h client.Hub, p client.Presence, isHealthy func() bool) *Server {
	baseCtx, cancelBase := context.WithCancel(context.Background())
	s := &Server{
		cfg:        cfg,
		verifier:   verifier,
		hub:        h,
		presence:   p,
		isHealthy:  isHealthy,
		baseCtx:    baseCtx,
		cancelBase: cancelBase,
	}
	// no Read/WriteTimeout: they would put deadlines on hijacked socket connections
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Service.APIBindAddress, cfg.Service.APIPort),
		Handler:           s.handler(),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return s
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()

	var health http.Handler = http.HandlerFunc(s.health)
	health = middlewares.RequestIDMiddleware(health)
	health = middlewares.LoggingMiddleware(health)
	mux.Handle("GET /v1/health", health)

	// the logging middleware's ResponseWriter does not implement http.Hijacker,
	// which websocket.Accept needs, so the upgrade route stays unwrapped
	mux.HandleFunc("GET /realtime", s.upgrade)

	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	status, natsStatus, code := "ok", "ok", http.StatusOK
	if !s.isHealthy() {
		status, natsStatus, code = "degraded", "unavailable", http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	body := map[string]any{"status": status, "checks": map[string]string{"nats": natsStatus}}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Errorf(r.Context(), "failed to write health response: %v", err)
	}
}

func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	userID, _ := s.verifier.UserID(r)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.cfg.WebSocket.AllowedOrigins})
	if err != nil {
		logger.Debugf(r.Context(), "websocket upgrade rejected: %v", err)
		return
	}

	// the request context must not be used after the connection is hijacked
	client.Serve(s.baseCtx, conn, userID, s.hub, s.presence)
}

func (s *Server) ListenAndServe() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown cancels every open socket first because http.Server.Shutdown does
// not track hijacked connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.cancelBase()
	return s.httpServer.Shutdown(ctx)
}
```

- [ ] **Step 3: Verify**

Run: `cd backend/realtime && go build ./... && go vet ./...`
Expected: no output, exit 0.

- [ ] **Step 4: Checkpoint.** Leave uncommitted.

---

### Task 7: Bootstrap and image

**Files:**
- Create: `backend/realtime/cmd/main.go`, `backend/realtime/Dockerfile`

**Interfaces:**
- Consumes: everything from Tasks 1–6.

- [ ] **Step 1: Create `backend/realtime/cmd/main.go`**

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/realtime/api"
	"sen1or/letslive/realtime/auth"
	cfg "sen1or/letslive/realtime/config"
	"sen1or/letslive/realtime/hub"
	"sen1or/letslive/realtime/presence"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/natsconn"
	"sen1or/letslive/shared/pkg/realtime"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"
)

var (
	configServiceName = "realtime_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	shutdownTimeout     = 15 * time.Second
	presenceGracePeriod = 5 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Init(logger.LogLevel(logger.Debug))

	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf(ctx, "failed to start discovery mechanism: %s", err)
	}

	cfgManager, err := sharedconfig.NewConfigManager[cfg.Config](ctx, registry, configServiceName, configProfile, cfg.PostProcess)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}
	config := cfgManager.GetConfig()

	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go sharedutils.RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config.Service.Hostname, config.Service.APIPort)

	otelShutdownFunc, err := tracer.SetupOTelSDK(ctx, *config)
	if err != nil {
		logger.Panicf(ctx, "failed to setup otel sdk: %v", err)
	}

	natsConn, err := natsconn.Connect(ctx, config.NATS.URL)
	if err != nil {
		logger.Panicf(ctx, "failed to connect to nats: %v", err)
	}

	publisher := realtime.NewNATSPublisher(natsConn)
	server := api.NewServer(
		config,
		auth.NewVerifier(config.AccessTokenSecret),
		hub.New(hub.NewNATSSource(natsConn), publisher),
		presence.NewTracker(publisher, presenceGracePeriod),
		natsConn.IsConnected,
	)

	go func() {
		logger.Infof(ctx, "starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		if err := server.ListenAndServe(); err != nil {
			logger.Errorf(ctx, "server stopped: %v", err)
		}
		stop()
	}()

	<-ctx.Done()
	logger.Infof(ctx, "shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	var shutdownWg sync.WaitGroup
	shutdownWg.Go(func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Errorf(shutdownCtx, "server shutdown: %v", err)
		}
	})
	shutdownWg.Go(func() {
		sharedutils.DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
	})
	shutdownWg.Go(func() {
		if err := otelShutdownFunc(shutdownCtx); err != nil {
			logger.Errorf(shutdownCtx, "otel shutdown: %v", err)
		}
	})
	shutdownWg.Go(func() {
		if err := natsConn.Drain(); err != nil {
			logger.Errorf(shutdownCtx, "failed to drain nats connection: %v", err)
		}
	})
	shutdownWg.Wait()

	logger.Infof(shutdownCtx, "service shut down complete.")
}
```

(`sync.WaitGroup.Go` exists since Go 1.25; the workspace is on 1.26.)

- [ ] **Step 2: Create `backend/realtime/Dockerfile`**

```dockerfile
# STAGE 1
FROM golang:1.26.5-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY realtime/go.mod realtime/go.sum ./realtime/
RUN go work init ./realtime
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/realtime
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY realtime/ ./realtime/

WORKDIR /usr/src/app/realtime
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
```

- [ ] **Step 3: Verify the whole workspace**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/backend
for m in shared user realtime; do (cd $m && go build ./... && go vet ./...) || echo "FAIL $m"; done
docker build -f realtime/Dockerfile -t letslive-realtime:dev .
docker build -f user/Dockerfile -t letslive-user:dev .
```

Expected: no `FAIL` lines, and both images build.

- [ ] **Step 4: Checkpoint.** Leave uncommitted.

---

### Task 8: Infrastructure and config profiles

**Files:**
- Modify: `docker-compose-dev.yaml`, `docker-compose.yaml`, `configs/kong.yml`
- Create: `../letslive-configs/realtime_service-dev.yml`, `../letslive-configs/realtime_service-prod.yml`
- Modify: `../letslive-configs/user_service-dev.yml`, `../letslive-configs/user_service-prod.yml`

- [ ] **Step 1: Add `realtime` to `docker-compose-dev.yaml`**

Insert after the `finance:` service block:

```yaml
  realtime:
    build:
      context: ./backend/
      dockerfile: realtime/Dockerfile
    container_name: letslive-realtime
    restart: always
    ports:
      - "7785:7785"
    expose:
      - "7785"
    environment:
      - CONFIG_SERVER_PROFILE=${CONFIG_SERVER_PROFILE}
      - CONFIG_SERVER_INTERVAL=${CONFIG_SERVER_INTERVAL}
      - REGISTRY_SERVICE_ADDRESS=${REGISTRY_SERVICE_ADDRESS}
      - ACCESS_TOKEN_SECRET=${ACCESS_TOKEN_SECRET}
    networks:
      general_network:
    depends_on:
      consul:
        condition: service_healthy
      nats:
        condition: service_healthy
```

In the `user:` service's `depends_on`, add:

```yaml
      nats:
        condition: service_healthy
```

- [ ] **Step 2: Same in `docker-compose.yaml`**

Use the identical block, but with `image: sen1or/letslive-realtime:latest` in place of the `build:` section. Add the same `nats` dependency to `user`.

- [ ] **Step 3: Add the Kong service in `configs/kong.yml`**

Insert after the `Chat_WS` service (before `- name: Finance`):

```yaml
  - name: Realtime
    host: realtime.service.consul
    port: 7785
    path: /
    connect_timeout: 60000
    read_timeout: 60000
    write_timeout: 60000
    routes:
      - name: Realtime_Socket
        paths:
          - /realtime
        strip_path: false
        protocols:
          - http
          - https
```

There's no `jwt` plugin on purpose (anonymous sockets). The gateway verifies the cookie itself.

- [ ] **Step 4: Create `../letslive-configs/realtime_service-dev.yml`**

```yaml
service:
  name: "realtime"
  hostname: "realtime" # used for docker compose specially
  apiBindAddress: "0.0.0.0"
  apiPort: 7785

nats:
  url: "nats://nats:4222"

websocket:
  allowedOrigins:
    - "localhost:3000"
    - "localhost:5000"

tracer:
  endpoint: "otel-collector:4318"
  secure: false
  batchTimeout: 5000 # milli
```

- [ ] **Step 5: Create `../letslive-configs/realtime_service-prod.yml`**

```yaml
service:
  name: realtime
  hostname: realtime
  apiBindAddress: 0.0.0.0
  apiPort: 7785

nats:
  url: nats://nats:4222

websocket:
  allowedOrigins:
    - letslive.work

tracer:
  endpoint: otel-collector:4318
  secure: false
  batchTimeout: 5000
```

- [ ] **Step 6: Add NATS to both user profiles**

Append to `../letslive-configs/user_service-dev.yml`:

```yaml

nats:
  url: "nats://nats:4222"
```

Append to `../letslive-configs/user_service-prod.yml`:

```yaml

nats:
  url: nats://nats:4222
```

- [ ] **Step 7: Ask before pushing `letslive-configs`**

The config server clones `CONFIG_SERVER_GIT_URI` (GitHub, branch `main`) at start, so the running stack only sees these files once they are pushed. Pushing is outward-facing: **show the owner `git -C ../letslive-configs diff` and the new files, and push only after they say so.** Until then, Task 10 cannot start.

- [ ] **Step 8: Verify the compose files parse**

Run: `docker compose -f docker-compose-dev.yaml config --quiet && docker compose -f docker-compose.yaml config --quiet`
Expected: no output, exit 0.

- [ ] **Step 9: Checkpoint.** Leave uncommitted.

---

### Task 9: Web: single socket provider and notification push

**Files:**
- Modify: `web/global.ts`
- Create: `web/constant/realtime.ts`, `web/types/realtime.ts`, `web/lib/realtime/parse-frame.ts`, `web/contexts/realtime-context.tsx`, `web/hooks/use-notification-realtime.ts`
- Modify: `web/hooks/queries/use-notifications.ts`, `web/app/(main)/_components/header/notification-bell.tsx`, `web/app/(main)/layout.tsx`

**Interfaces:**
- Produces:
  - `GLOBAL.REALTIME_URL`
  - `RealtimeProvider`; `useRealtime(): { onEvent(type, handler) => unsubscribe; onReconnect(handler) => unsubscribe }`
  - `useNotificationRealtime(enabled: boolean)`
  - `NOTIFICATIONS_ROOT_QUERY_KEY`

Phases 2–3 add `subscribe(topic)` to `useRealtime`. Don't add it now.

- [ ] **Step 1: Read the Next 16 notes**

Run: `ls web/node_modules/next/dist/docs/ && grep -ril "use client" web/node_modules/next/dist/docs | head`

Skim the client-component and layout guides for anything that changes how a `"use client"` context provider is mounted inside a server layout. (`web/AGENTS.md` requires this before writing Next code.) If nothing differs from React's `createContext` + provider pattern, continue.

- [ ] **Step 2: Collapse the URL builders in `web/global.ts`**

Replace `getWebSocketUrl` and `getDmWebSocketUrl` (lines 30–83) with one builder:

```ts
function getWsUrl(path: string): string {
    const wsProtocol = process.env.NEXT_PUBLIC_BACKEND_WS_PROTOCOL?.trim();
    const ipAddress = process.env.NEXT_PUBLIC_BACKEND_IP_ADDRESS?.trim();
    const port = process.env.NEXT_PUBLIC_BACKEND_PORT?.trim();

    if (
        !wsProtocol ||
        !ipAddress ||
        !port ||
        wsProtocol === "" ||
        ipAddress === "" ||
        port === ""
    ) {
        if (typeof window !== "undefined") {
            console.error("Missing or empty WebSocket environment variables:", {
                wsProtocol: wsProtocol || "(empty or undefined)",
                ipAddress: ipAddress || "(empty or undefined)",
                port: port || "(empty or undefined)",
            });
        }
        throw new Error(
            "Missing or empty required environment variables: NEXT_PUBLIC_BACKEND_WS_PROTOCOL, NEXT_PUBLIC_BACKEND_IP_ADDRESS, NEXT_PUBLIC_BACKEND_PORT",
        );
    }

    return `${wsProtocol}://${ipAddress}:${port}${path}`;
}

const GLOBAL = Object.freeze({
    API_URL: getBackendUrl(),
    WS_SERVER_URL: getWsUrl("/ws"),
    DM_WS_SERVER_URL: getWsUrl("/dm-ws"),
    REALTIME_URL: getWsUrl("/realtime"),
});
```

The URLs for `WS_SERVER_URL` and `DM_WS_SERVER_URL` are unchanged.

- [ ] **Step 3: Create `web/constant/realtime.ts`**

```ts
export const REALTIME_EVENT = {
    NOTIFICATION_CREATED: "notification.created",
} as const;

export const REALTIME_RECONNECT_INITIAL_DELAY_MS = 1000;
export const REALTIME_RECONNECT_MAX_DELAY_MS = 30_000;
```

- [ ] **Step 4: Create `web/types/realtime.ts`**

```ts
export type RealtimeEventFrame = {
    op: "event";
    topic: string;
    type: string;
    data: unknown;
};

export type RealtimeErrorFrame = {
    op: "error";
    code: string;
};

export type RealtimePongFrame = {
    op: "pong";
};

export type RealtimeServerFrame =
    | RealtimeEventFrame
    | RealtimeErrorFrame
    | RealtimePongFrame;
```

- [ ] **Step 5: Create `web/lib/realtime/parse-frame.ts`**

```ts
import type { RealtimeServerFrame } from "@/types/realtime";

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
}

export function parseRealtimeFrame(raw: unknown): RealtimeServerFrame | null {
    if (typeof raw !== "string") return null;

    let parsed: unknown;
    try {
        parsed = JSON.parse(raw);
    } catch {
        return null;
    }
    if (!isRecord(parsed)) return null;

    switch (parsed.op) {
        case "event":
            return typeof parsed.topic === "string" &&
                typeof parsed.type === "string"
                ? {
                      op: "event",
                      topic: parsed.topic,
                      type: parsed.type,
                      data: parsed.data,
                  }
                : null;
        case "error":
            return typeof parsed.code === "string"
                ? { op: "error", code: parsed.code }
                : null;
        case "pong":
            return { op: "pong" };
        default:
            return null;
    }
}
```

- [ ] **Step 6: Create `web/contexts/realtime-context.tsx`**

```tsx
"use client";

import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useRef,
    type ReactNode,
} from "react";
import GLOBAL from "@/global";
import useUser from "@/hooks/user";
import { parseRealtimeFrame } from "@/lib/realtime/parse-frame";
import type { RealtimeEventFrame } from "@/types/realtime";
import {
    REALTIME_RECONNECT_INITIAL_DELAY_MS,
    REALTIME_RECONNECT_MAX_DELAY_MS,
} from "@/constant/realtime";

type EventHandler = (frame: RealtimeEventFrame) => void;
type Unsubscribe = () => void;

export type RealtimeContextValue = {
    onEvent: (type: string, handler: EventHandler) => Unsubscribe;
    onReconnect: (handler: () => void) => Unsubscribe;
};

const RealtimeContext = createContext<RealtimeContextValue | null>(null);

export function RealtimeProvider({ children }: { children: ReactNode }) {
    const userId = useUser((state) => state.user?.id ?? null);
    const eventHandlersRef = useRef(new Map<string, Set<EventHandler>>());
    const reconnectHandlersRef = useRef(new Set<() => void>());

    // userId is a dependency so login/logout reopens the socket with the new
    // cookie identity
    useEffect(() => {
        let disposed = false;
        let hasOpened = false;
        let delay = REALTIME_RECONNECT_INITIAL_DELAY_MS;
        let socket: WebSocket | null = null;
        let retryTimer: ReturnType<typeof setTimeout> | null = null;

        const connect = () => {
            const ws = new WebSocket(GLOBAL.REALTIME_URL);
            socket = ws;

            ws.onopen = () => {
                delay = REALTIME_RECONNECT_INITIAL_DELAY_MS;
                if (hasOpened) {
                    reconnectHandlersRef.current.forEach((handler) =>
                        handler(),
                    );
                }
                hasOpened = true;
            };

            ws.onmessage = (message: MessageEvent<unknown>) => {
                const frame = parseRealtimeFrame(message.data);
                if (frame?.op === "event") {
                    eventHandlersRef.current
                        .get(frame.type)
                        ?.forEach((handler) => handler(frame));
                } else if (frame?.op === "error") {
                    console.error("[realtime] server error:", frame.code);
                }
            };

            ws.onclose = () => {
                if (disposed || socket !== ws) return;
                retryTimer = setTimeout(connect, delay);
                delay = Math.min(delay * 2, REALTIME_RECONNECT_MAX_DELAY_MS);
            };
        };

        connect();

        return () => {
            disposed = true;
            if (retryTimer) clearTimeout(retryTimer);
            socket?.close();
        };
    }, [userId]);

    const onEvent = useCallback(
        (type: string, handler: EventHandler): Unsubscribe => {
            const handlers = eventHandlersRef.current;
            const forType = handlers.get(type) ?? new Set<EventHandler>();
            forType.add(handler);
            handlers.set(type, forType);
            return () => {
                forType.delete(handler);
                if (forType.size === 0 && handlers.get(type) === forType) {
                    handlers.delete(type);
                }
            };
        },
        [],
    );

    const onReconnect = useCallback((handler: () => void): Unsubscribe => {
        reconnectHandlersRef.current.add(handler);
        return () => {
            reconnectHandlersRef.current.delete(handler);
        };
    }, []);

    const value = useMemo(
        () => ({ onEvent, onReconnect }),
        [onEvent, onReconnect],
    );

    return (
        <RealtimeContext.Provider value={value}>
            {children}
        </RealtimeContext.Provider>
    );
}

export function useRealtime(): RealtimeContextValue {
    const context = useContext(RealtimeContext);
    if (!context) {
        throw new Error("useRealtime must be used within RealtimeProvider");
    }
    return context;
}
```

- [ ] **Step 7: Drop polling in `web/hooks/queries/use-notifications.ts`**

Add a root key above the existing keys:

```ts
export const NOTIFICATIONS_ROOT_QUERY_KEY = ["notifications"] as const;
```

In `useUnreadNotificationCount`, delete these two lines:

```ts
        refetchInterval: 30_000,
        refetchIntervalInBackground: false,
```

- [ ] **Step 8: Create `web/hooks/use-notification-realtime.ts`**

```ts
"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useRealtime } from "@/contexts/realtime-context";
import { REALTIME_EVENT } from "@/constant/realtime";
import { NOTIFICATIONS_ROOT_QUERY_KEY } from "@/hooks/queries/use-notifications";

export default function useNotificationRealtime(enabled: boolean) {
    const { onEvent, onReconnect } = useRealtime();
    const queryClient = useQueryClient();

    useEffect(() => {
        if (!enabled) return;

        const refresh = () => {
            queryClient.invalidateQueries({
                queryKey: NOTIFICATIONS_ROOT_QUERY_KEY,
            });
        };
        const offEvent = onEvent(REALTIME_EVENT.NOTIFICATION_CREATED, refresh);
        const offReconnect = onReconnect(refresh);

        return () => {
            offEvent();
            offReconnect();
        };
    }, [enabled, onEvent, onReconnect, queryClient]);
}
```

Invalidating `["notifications"]` refreshes both the list (`["notifications","list"]`) and the unread count (`["notifications","unread-count"]`). This matches what the 30s poll refreshed before, plus the list.

- [ ] **Step 9: Use it in `web/app/(main)/_components/header/notification-bell.tsx`**

Add the import:

```ts
import useNotificationRealtime from "@/hooks/use-notification-realtime";
```

Right after `const unreadCount = unreadData?.count ?? 0;`, add:

```ts
    useNotificationRealtime(!!user);
```

- [ ] **Step 10: Mount the provider in `web/app/(main)/layout.tsx`**

```tsx
import { Header } from "@/app/(main)/_components/header/header";
import { MainBodyLayout } from "@/app/(main)/_components/main-body-layout";
import { RealtimeProvider } from "@/contexts/realtime-context";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <RealtimeProvider>
            <div className="flex h-screen w-screen flex-col overflow-hidden">
                <Header />
                <MainBodyLayout>{children}</MainBodyLayout>
            </div>
        </RealtimeProvider>
    );
}
```

- [ ] **Step 11: Verify**

Run: `cd web && npx tsc --noEmit && npm run lint`
Expected: no type errors and no lint errors. If the full lint already reports problems in files this task did not touch, run `npx eslint` on just the touched files and require those to be clean.

- [ ] **Step 12: Checkpoint.** Leave uncommitted.

---

### Task 10: Running-stack verification

Prerequisite: the owner has pushed `letslive-configs` (Task 8 step 7). The web dev server runs on the owner's `:3000`; don't start another one on that port.

- [ ] **Step 1: Bring up the changed services**

```bash
cd /Users/ictsaigon.vn/mywork/letslive
docker compose -f docker-compose-dev.yaml up -d --build realtime user
docker compose -f docker-compose-dev.yaml restart kong
docker compose -f docker-compose-dev.yaml logs --tail=50 realtime
curl -s localhost:7785/v1/health
```

Expected: the logs show `starting server on realtime:7785`, and health returns `{"checks":{"nats":"ok"},"status":"ok"}`. Check Consul too: `curl -s localhost:8500/v1/health/service/realtime | grep -o '"Status":"[a-z]*"'` shows `passing`.

- [ ] **Step 2: Anonymous socket through Kong**

In the browser at `http://localhost:3000` (logged out), open DevTools → Network → WS. Expect a `ws://localhost:8000/realtime` connection with status 101 and no frames. In the console: no `[realtime]` errors.

- [ ] **Step 3: Origin check (Review Focus 1)**

```bash
curl -s -o /dev/null -w "%{http_code}\n" -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  -H "Origin: http://evil.example" localhost:8000/realtime
```

Expected: `403`. The same command with `-H "Origin: http://localhost:3000"` returns `101`. Abort it with Ctrl-C, since it holds the socket open.

- [ ] **Step 4: Bad cookie is anonymous, not rejected (Review Focus 2)**

In one terminal, watch presence events:

```bash
docker run --rm -it --network letslive_general_network natsio/nats-box:latest \
  nats --server nats://nats:4222 sub rt.presence
```

(If the network name differs, find it with `docker network ls | grep general`.) In another terminal, repeat the allowed-origin command from Step 3 with `-H "Cookie: ACCESS_TOKEN=forged.token.value"`. Expected: `101`, and **no** message on `rt.presence`, because the socket is anonymous. For contrast, logging in at `:3000` produces `{"userId":"<id>","online":true}`.

`natsio/nats-box` is only a throwaway CLI container; it's not a project dependency.

- [ ] **Step 5: Notification push replaces polling**

Log in at `:3000`. Note your user id (`GET /v1/user/me` in the Network tab). Then:

```bash
curl -s -X POST localhost:7778/v1/notifications -H 'Content-Type: application/json' \
  -d '{"userId":"<your-id>","type":"gift_received","title":"Realtime check","message":"pushed over the gateway"}'
```

Expected:
- the bell badge increments within about 1s
- the WS frame shows `{"op":"event","topic":"user:<id>","type":"notification.created","data":{...}}`
- the Network tab shows one `unread-count` refetch right after the frame, and no request every 30s while idle (watch for 70s)

Also send a gift through the UI. The recipient's bell updates the same way, which confirms the `GiftService` path.

- [ ] **Step 6: Two tabs (Review Focus 3)**

Open a second tab as the same user. Send the curl from Step 5 again: both badges update. Close one tab and send it again: the remaining tab still updates.

- [ ] **Step 7: Malformed publish is dropped (Review Focus 4)**

```bash
docker run --rm --network letslive_general_network natsio/nats-box:latest \
  nats --server nats://nats:4222 pub "rt.user.<your-id>" 'not json'
```

Expected: the realtime logs show `dropping malformed realtime message on user:<id>`. Then repeat Step 5: the bell still updates.

- [ ] **Step 8: NATS restart (Review Focus 5)**

With the tab open, run `docker compose -f docker-compose-dev.yaml restart nats`, wait about 10s, then repeat Step 5. Expected: the bell updates without restarting `realtime` or reloading the page.

- [ ] **Step 9: Reconnect refetch**

Run `docker compose -f docker-compose-dev.yaml restart realtime`. Expected: the WS connection closes, a new one opens (after 1s and 2s backoff), and the Network tab shows one notification refetch right after it opens.

- [ ] **Step 10: Old paths untouched**

Open a live chat (`/users/<id>`) and `/messages`. Both still use `/ws` and `/dm-ws` and work exactly as before this change.

- [ ] **Step 11: Hand back for review**

Report each step's result to the owner, including the parity notes from Task 2. Leave everything uncommitted until they approve. After approval, commit with message refs `docs/superpowers/specs/2026-10-06-realtime-gateway-design.md`, push `letslive-configs` separately if not already pushed, and open the PR to `main`.
