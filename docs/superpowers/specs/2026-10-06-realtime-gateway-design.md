# Realtime Gateway + Go Chat Service Design

**Date:** 2026-10-06
**Author:** ThNam203

---

## Overview

This design does two things together:

1. **Realtime gateway.** Move every WebSocket out of `chat` into a new Go service, `realtime`. It is a push-only transport: it authenticates sockets, tracks topic subscriptions and fans out events that domain services publish over core NATS. It holds no business rules. All client writes go over REST to the owning service.
2. **Chat in Go.** Port `chat` from Node/Express to Go so the backend uses one language. The Go port carries only REST + NATS publishing; the socket code is deleted rather than ported, because the gateway replaces it.

Today the Node `chat` serves two sockets (`/ws` live chat, `/dm-ws` DMs) in the same process as its REST API, with socket handling and persistence interleaved. Notifications and DM unread counts are polled every 30s. Future realtime features (notifications, gifts, viewer counts, presence) need one shared socket instead of one per feature.

---

## Decisions Made

| Question | Decision |
|---|---|
| Gateway language | Go, new module `backend/realtime`, same bootstrap as other Go services |
| Gateway port / Kong path | `7785` (7784 is `admin_service`) / `/realtime` |
| Service → gateway transport | Core NATS (fire-and-forget). Not JetStream: data is persisted before publish, pushes are hints, replay/acks add cost with no benefit |
| Client → server writes | All REST (Discord model). The socket accepts only `subscribe` / `unsubscribe` / `ping` |
| Instances | Single gateway instance. Presence is in-memory. Fan-out already goes through NATS, so scaling out later only needs presence moved to a shared store |
| Chat language | Ported to Go, merged into migration phases 2–4. Kong moves path groups from Node to Go one phase at a time. Node is deleted in phase 4 |
| Naming during migration | `git mv backend/chat backend/chat-legacy`. The Go module takes `backend/chat` and the Consul name `chat`. Node runs as `chat-legacy` until deleted |
| Redis | Removed entirely in phase 4 (only Node chat uses it, only for socket concerns) |
| Existing bugs (impersonation, un-awaited membership check, leaked ping timer, one socket per user) | Out of scope; fixed later. Most disappear because the Node socket code is deleted, not ported |
| JetStream later | Kept open: subjects are stable and publishing/subscribing sit behind interfaces (see "Future: JetStream") |

---

## Architecture

```
browser ──1 socket──► Kong /realtime ──► realtime (Go, :7785)
   │                                        ▲ core NATS subscribe rt.user.<id> / rt.room.<id>
   │                                        │
   └──REST (Kong JWT)──► user / chat (Go) ──persist──► Postgres / Mongo
                              └──core NATS publish──┘
realtime ──publish rt.presence──► chat (computes DM contacts, republishes to their user topics)
```

### Topics

| Client topic | NATS subject | Who receives | How subscribed |
|---|---|---|---|
| `user:<userId>` | `rt.user.<userId>` | that user only | automatically on an authenticated socket; clients cannot subscribe to it explicitly |
| `room:<roomId>` | `rt.room.<roomId>` | anyone, including anonymous | explicit `subscribe` |

Private data is always delivered to `user:` topics (services already know recipient ids), so the gateway never has to authorize a subscription beyond "is this a `room:` topic".

Topic ids must match `^[A-Za-z0-9-]{1,36}$`. This blocks NATS wildcard/token injection (`*`, `>`, `.`).

### Wire protocol (JSON text frames)

Client → server:

```json
{"op":"subscribe","topic":"room:<id>"}
{"op":"unsubscribe","topic":"room:<id>"}
{"op":"ping"}
```

Server → client:

```json
{"op":"event","topic":"user:<id>","type":"notification.created","data":{...}}
{"op":"pong"}
{"op":"error","code":"invalid_topic|forbidden_topic|too_many_subscriptions|bad_frame"}
```

The NATS message body that services publish is `{"type": "...", "data": {...}}`. The gateway wraps it with `op` and `topic`. It never inspects `data`.

### Authentication

- The Kong `/realtime` route has **no** JWT plugin, because anonymous viewers must connect for live chat.
- Because Kong is not verifying, the gateway itself verifies the `ACCESS_TOKEN` cookie signature (HS256, `ACCESS_TOKEN_SECRET` env var, the same secret `auth` signs with), plus `exp`. This differs from other Go services, which `ParseUnverified` behind Kong.
- Valid token → authenticated socket, auto-subscribed to `user:<userId>`.
- Missing or invalid token → anonymous socket, `room:` topics only. The upgrade is not rejected.
- Token expiry mid-connection is not enforced (same as today). The socket keeps the identity it connected with.

### Limits and keep-alive

| Limit | Value |
|---|---|
| Inbound frame size | 1 KiB (larger → the library closes the socket with status 1009, message too big) |
| Allowed origins | `websocket.allowedOrigins` in the config profile (dev: `localhost:3000`, `localhost:5000`; prod: `letslive.work`). Other origins get 403. Because sockets authenticate with a cookie, this is the cross-site WebSocket hijacking defence |
| Subscriptions per socket | 50 |
| Per-socket outbound buffer | 64 frames; if full, the socket is closed (slow consumer) and the client reconnects and refetches |
| Write timeout | 10s |
| Server ping | every 30s (Kong closes idle upstreams at 60s) |

### Presence

- The gateway keeps an in-memory count of open authenticated sockets per user.
- 0→1 publishes `{"userId":..., "online":true}` on NATS subject `rt.presence`.
- 1→0 waits a 5s grace period. If the user reconnects within it, no event is sent. Otherwise it publishes `online:false`.
- `chat` subscribes to `rt.presence`, computes the user's DM contacts (participants of their conversations, as the Node `broadcastPresence` does today) and publishes `dm:user_online` / `dm:user_offline` to each contact's `rt.user.<id>`.
- Behaviour change: a user is "online" on any page while logged in, not only while on `/messages`.

### Room membership events (live chat join/leave lines)

- The gateway counts authenticated subscribers per (user, room).
- On 0→1 it publishes `member.joined` to that room's NATS subject, and on 1→0 it publishes `member.left`. The payload is `{userId}`.
- Anonymous subscribers produce no events.
- These are generic topic-membership events, not chat-specific. The web resolves the username with the existing `usePublicUser(userId)` query.

---

## Gateway Components (`backend/realtime`)

| Unit | Purpose | Depends on |
|---|---|---|
| `cmd/main.go` | Bootstrap: Consul, config server (`realtime_service`), OTel, NATS connect, HTTP server, graceful shutdown | everything below |
| `config/` | `Config` struct (service, nats url, tracer) | shared config |
| `auth/` | `UserIDFromRequest(r) (userId string, ok bool)`: cookie → verified HS256 claims | `golang-jwt/jwt/v5` (already used) |
| `protocol/` | Frame types, `ParseClientFrame`, `ParseTopic` → `(kind, id)`, topic ↔ subject mapping | none |
| `hub/` | Topic registry: local subscribers per topic, ref-counted NATS subscription per topic through a `Source` interface, fan-out, membership events | `Source` |
| `hub/natssource.go` | `Source` implementation over core NATS (`Subscribe` / `Unsubscribe`) | `nats.go` (already in shared) |
| `presence/` | Per-user socket count, grace timer (injectable clock), publishes via `Publisher` | `shared/pkg/realtime.Publisher` |
| `client/` | One socket: read loop (parse, limits), write loop (buffered channel, ping, write timeout), cleanup | hub, presence, protocol |
| `api/` | `GET /v1/health`, `GET /realtime` upgrade | client, auth |

`shared/pkg/realtime` (new) provides:
- a `Publisher` interface: `Publish(ctx, topic string, eventType string, data any) error`
- a core-NATS implementation
- subject helpers (`UserTopic(id)`, `RoomTopic(id)`, `SubjectFor(topic)`)
- a NATS connect helper with retry/backoff (same shape as `natsbus.connect`)

Go services only call `Publisher`.

---

## Chat Service in Go (`backend/chat`)

Same layout and bootstrap as `user` / `finance`:
- `cmd/main.go`, `config/`, `api/`, `handlers/`, `services/`, `repositories/`, `domains/`, `dto/`, `response/`, `gateway/`
- shared logger, tracer, discovery, middlewares, `shared/response`
- config server profile `chat_service-{dev,prod}.yml`, port `7780`
- user id from the `ACCESS_TOKEN` cookie via `ParseUnverified` behind Kong, like the other services

### Behaviour parity

The Node implementation is the reference: `chat-legacy/src/services/*.ts`, `handlers/*.ts`, `models/*.ts`. Every endpoint keeps:
- its path
- its request validation limits
- its response envelope, codes and keys (`res_err_conversation_not_found` 50019 … `res_err_user_setup_incomplete` 50026, carried into `response/` templates)
- its JSON field names

The Go handlers are ported from the Node logic rule by rule. They are not redesigned.

### Mongo compatibility (Node and Go share `chat_db` during phases 2–3)

| Collection | Model |
|---|---|
| `conversations` | `Conversation`: `participants[]`, `lastMessage`, `createdAt` / `updatedAt` |
| `dmmessages` | `DmMessage`: `conversationId` ObjectId, `readBy[]`, `replyTo` ObjectId, `isDeleted`, timestamps |
| `messages` | `Message`: live chat `roomId`, `userId`, `username`, `text`, `timestamp` |
| `chat_commands` | `ChatCommand`: unique `(scope, ownerId, name)`, `name` lowercased and trimmed |

- Field names and BSON types are identical to what Mongoose writes. Ids are `ObjectId`. Dates are BSON dates. Inserts set `__v: 0`. The Go code maintains `createdAt` / `updatedAt` itself on insert and update.
- Mongoose applies schema defaults (e.g. `profilePicture: null`, `isMuted: false`, `imageUrls: []`, `readBy: []`, `lastMessage: null`). Go structs write them explicitly, so documents look the same whichever service created them.
- Indexes: Go runs `createIndexes` at startup with the same key specs and options as the Mongoose schemas. Identical specs are no-ops on an existing database.
- JSON output: ObjectIds as hex strings, dates as ISO-8601 with milliseconds in UTC (`2026-10-06T10:00:00.000Z`), matching Mongoose `toJSON`.
- Shape check: a document written by Go and one written by Node are compared field by field (see Verification).

### New in Go (beyond parity)

- Publishing through `shared/pkg/realtime.Publisher` after successful writes (see Feature Flows).
- `POST /v1/conversations/:id/typing`
- `POST /v1/messages` (live chat send)
- A `rt.presence` subscriber (core NATS) that fans presence out to DM contacts.

---

## Feature Flows and Migration Phases

Each phase leaves the app fully working. Kong decides which service serves each path, so every phase can be rolled back by switching a route back.

### Phase 1: Gateway + notifications (chat untouched)

1. `shared/pkg/realtime` + `backend/realtime`.
2. `user` service: after `NotificationService.CreateNotification` succeeds, publish `notification.created` to `user:<recipientId>` with the created notification DTO. A publish failure is logged and does not fail the request (the notification row exists, and the client refetches on reconnect).
3. Web: a new `RealtimeProvider`, mounted in the `(main)` layout for every visitor, owns the single socket. It reconnects with exponential backoff (1s → 30s) and exposes `subscribe(topic, handler)` and `onEvent(type, handler)`.
4. Notification bell: on `notification.created`, invalidate the list query and the unread-count query. Remove the 30s `refetchInterval` and invalidate both on socket reconnect.
5. Infra:
   - `realtime` in `docker-compose.yaml` and `docker-compose-dev.yaml`
   - Kong service `Realtime` → `realtime.service.consul:7785`, route `/realtime`
   - `nats.url` in the `user_service` and `realtime_service` config profiles (same place as every other service setting); `ACCESS_TOKEN_SECRET` env for `realtime`
   - `realtime_service-{dev,prod}.yml` in `letslive-configs`
   - `GLOBAL.REALTIME_URL` in web

### Phase 2: Go chat with DMs

1. Rename:
   - `git mv backend/chat backend/chat-legacy`
   - Node registers in Consul as `chat-legacy`, with hostname `chat-legacy` and internal port 7780; the host port mapping moves to `7786:7780`
   - compose service renamed `chat-legacy`
2. Create the Go `backend/chat` with:
   - health
   - all conversation routes (`GET/POST /v1/conversations`, `GET /v1/conversations/unread-counts`, `GET/PUT/DELETE /v1/conversations/:id`, `POST /v1/conversations/:id/participants`, `DELETE /v1/conversations/:id/participants/:userId`)
   - all DM message routes (`GET/POST /v1/conversations/:id/messages`, `PATCH/DELETE /v1/conversations/:id/messages/:msgId`, `POST /v1/conversations/:id/read`)
   - new `POST /v1/conversations/:id/typing`
   - the presence subscriber
3. Kong:
   - service `Chat` → `chat.service.consul:7780` (Go) serves `/conversations`
   - new service `Chat_Legacy` → `chat-legacy.service.consul:7780` keeps `/messages`, `/chat-commands`
   - `Chat_WS` points at `chat-legacy` and keeps `/ws` only (`/dm-ws` removed)
4. DM flows:

| Action | Before (Node socket) | After (Go REST + gateway) |
|---|---|---|
| Send | `dm:send_message` frame | `POST /v1/conversations/:id/messages`. On success, publish `dm:new_message` to every participant's `user:` topic (sender included, for multi-tab). Errors come back in the REST response; `dm:send_failed` is retired |
| Typing | `dm:typing_start/stop` frames | `POST /v1/conversations/:id/typing` `{ "state": "start" \| "stop" }`. The client sends `start` at most once per 3s while typing. Chat checks participation and publishes `dm:user_typing` / `dm:user_stopped_typing` to the other participants |
| Mark read | `dm:mark_read` frame, plus a REST call already made by the page | `POST /v1/conversations/:id/read`. On success, publish `dm:read_receipt` to the other participants |
| Presence | DM socket connect/close | gateway `rt.presence` → chat → contacts |

5. Web:
   - event `data` keeps today's DM payload shapes (minus `type` and `recipientIds`), so the current event handling is reused
   - `use-dm-websocket` is rewritten as `use-dm-realtime` on top of `RealtimeProvider`
   - `DmWebSocketProvider` and `DM_WS_SERVER_URL` usage are removed
   - DM unread-count polling is removed; `dm:new_message` already increments it, and reconnect invalidates it

Edit/delete pushes (`dm:message_edited` / `dm:message_deleted`) are not emitted today and stay out of scope.

### Phase 3: Live chat + chat commands in Go

1. Go `chat` adds:
   - `GET /v1/messages?roomId=` (latest 50, oldest first, same as today)
   - new `POST /v1/messages` `{ roomId, text }`
   - chat commands (`GET /v1/chat-commands`, `GET /v1/chat-commands/mine`, `POST /v1/chat-commands`, `PATCH/DELETE /v1/chat-commands/:id`)
2. Kong: `/messages` and `/chat-commands` move to `Chat` (Go). The `/messages` route is split by method: `GET` is public, `POST` gets the JWT plugin. `Chat_Legacy` keeps only `Chat_WS` `/ws`, now unused.
3. Live chat flow:
   - Receive: the chat panel subscribes `room:<roomId>` on mount and unsubscribes on unmount. Anonymous viewers subscribe too (today they receive nothing because they never `JOIN`).
   - Send: `POST /v1/messages`. Chat takes `userId` from the cookie, resolves `username` through the user service's internal batch identity endpoint (as Node's `UserServiceGateway.getIdentities` does), saves to Mongo, then publishes `chat.message` (today's `ReceivedMessage` shape) to `room:<roomId>`.
   - Join/leave lines come from the gateway's `member.joined` / `member.left`.
4. Web: the chat panel uses `RealtimeProvider` and the REST send. `WS_SERVER_URL` usage is removed.

### Phase 4: Delete the old path

- Delete `backend/chat-legacy` entirely.
- Remove from both compose files: the `chat-legacy` service, the `chat_pubsub` Redis container and its volume, and the swagger entry for the chat spec (replaced by the Go chat OpenAPI).
- Kong: delete `Chat_Legacy` and `Chat_WS`.
- Web: delete `GLOBAL.WS_SERVER_URL`, `GLOBAL.DM_WS_SERVER_URL` and their URL builders, `dm-websocket-context.tsx`, `use-dm-websocket.ts`, and the unused client DM event types.
- Docs: `backend/chat/docs/openapi.yaml` for the Go REST API, and `backend/realtime/docs/openapi.yaml` describing the socket protocol.

---

## Error Handling

| Failure | Behaviour |
|---|---|
| NATS down at gateway/chat start | Connect with retry/backoff. The health check reports unhealthy until connected |
| NATS drops later | `nats.go` reconnects automatically and re-establishes subscriptions. Pushes during the gap are lost; clients are unaffected because data is in the DB |
| Publish fails in user/chat | Log with context. The REST request still succeeds, because the write is already persisted |
| Bad client frame | `{"op":"error","code":"bad_frame"}`. The socket stays open. Oversized frames close it |
| Slow consumer | Socket closed. The client reconnects and invalidates its queries |
| Client reconnect | Re-subscribe all `room:` topics it holds and invalidate notification / DM unread / open conversation queries |
| Mongo unavailable at chat start | Connect with retry; startup fails after the retry budget, like `sharedutils.ConnectDB` does for Postgres |

---

## Verification

Unit and integration tests are out of scope for this work; the owner writes them later. Each phase is verified with:

- **Static checks:** `go build ./...` and `go vet ./...` for every touched Go module; `npx tsc --noEmit` and `npm run lint` in `web`.
- **Parity review (phases 2–3):** each Go handler/service is checked rule by rule against its Node source (`chat-legacy/src/services/*.ts`, `handlers/*.ts`, `models/*.ts`): validation limits, error codes and keys, authorization checks, query filters, sort order, pagination, and JSON field names. The checklist is reported with the phase.
- **Mongo shape check (phases 2–3):** a document written by Go and one written by Node are compared field by field in the compose `chat_db` with `mongosh`.
- **Running stack, after each phase:**
  - Phase 1: a notification created through the user service updates the bell instantly, with no polling request in the network tab
  - Phase 2: DMs send, type, read, show presence, and create/leave conversations across two browsers
  - Phase 3: live chat works for a logged-in user and an anonymous viewer, and chat commands work
  - Phase 4: `docker compose` runs without `chat-legacy` and `chat_pubsub`

---

## Future: JetStream

Subjects `rt.user.>` and `rt.room.>` are stable. To add offline catch-up later:
- create a stream over `rt.user.>` with a short max age
- add a sequence number to `event` frames
- add a `{"op":"resume","topic":...,"after":<seq>}` client op
- swap `hub/natssource.go` for a JetStream ordered-consumer source

Publishers do not change. All of it is additive, so old clients keep working.

---

## New Dependencies (approved)

| Where | Package | Why |
|---|---|---|
| `backend/realtime` | `github.com/coder/websocket` | WebSocket server. Context-aware, safe for concurrent writes |
| `backend/chat` (Go) | `go.mongodb.org/mongo-driver/v2` | Official Mongo driver |

`nats.go` and `golang-jwt/jwt/v5` are already in the workspace. The `nats` npm package is no longer needed, because Node chat is never changed to publish.
