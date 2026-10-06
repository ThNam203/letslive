# Realtime Phase 2: Go Chat Service with DMs

**Date:** 2026-10-06
**Author:** ThNam203
**Builds on:** Phase 1 (`feat/realtime-gateway`, PR #366)

---

## Goal

Port the DM and conversation half of the Node `chat` service to a new Go service, move DM realtime onto the Phase 1 gateway, and stop using `/dm-ws`.

After this phase:

- Go `chat` serves every `/v1/conversations/**` route, plus the new `POST /v1/conversations/:id/typing`.
- DM pushes (new message, typing, read receipt, presence) go out over core NATS to the gateway, then to each user's `user:<id>` topic.
- The web sends DMs and typing over REST and receives everything on the single realtime socket.
- Node runs on as `chat-legacy` and still serves live chat (`/messages`, `/chat-commands`, `/ws`) until Phase 3.

## Design recap

The design spec from the original session was never committed, so the decisions this phase relies on are repeated here.

| Topic | Decision |
|---|---|
| Rename | `git mv backend/chat backend/chat-legacy`. Node registers in Consul as `chat-legacy` with hostname `chat-legacy`, internal port 7780, host port `7786:7780` |
| Go service | `backend/chat`, Consul name `chat`, port 7780, same bootstrap as `realtime` / `finance`, config profile `chat_service-{dev,prod}.yml` |
| Database | Go and Node share `chat_db` (Mongo, database `chat`) during phases 2–3. Go writes documents in exactly the shape Mongoose writes them |
| Driver | `go.mongodb.org/mongo-driver/v2` (v2.9.1) |
| Writes | All REST. The socket only pushes |
| Push transport | `shared/pkg/realtime.Publisher` (core NATS). A publish failure is logged and never fails the request, because the data is already saved |
| Presence | The gateway publishes `{userId, online}` on `rt.presence`. Go chat works out the user's DM contacts and pushes `dm:user_online` / `dm:user_offline` to each contact's `user:` topic |
| Event payloads | The `data` of each event keeps today's `/dm-ws` payload shape minus `type` and `recipientIds`, so the web's existing event handling is reused |
| Rollback | Kong decides which service serves `/conversations`. Pointing the `Chat` Kong service back at `chat-legacy` restores the Node REST API. Realtime DMs would stop in that case, but REST keeps working |

## Branch and working rules

- Stack a new branch, `feat/realtime-phase2-go-chat`, on `feat/realtime-gateway`, because PR #366 isn't merged yet. Rebase it onto `main` once #366 merges.
- Standing rules from you: no unit tests, no per-task commits, no worktree. Everything stays uncommitted until you've reviewed it.
- Verification is `go build` / `go vet` / `gofmt`, web `tsc` + lint, Docker image builds, and a parity read-through against the Node code. You run the checks against the running stack.

---

## Parity reference (Node → Go)

The Node code is the reference: `chat-legacy/src/handlers/conversationHandler.ts`, `handlers/dmMessageHandler.ts`, `services/conversationService.ts`, `services/dmMessageService.ts`, `models/*.ts`, `types/api-response.ts`. Each rule below must hold in Go. The "Node" column is what happens today.

### Auth

- The user id comes from the `ACCESS_TOKEN` cookie, read with an unverified parse (Kong already checked the signature). Node also rejects a token whose `exp` has passed, so Go checks `exp` too. `jwt.ParseUnverified` does not.
- Missing, unparseable or expired token, or no `userId` claim: `401 res_err_unauthorized` (20005).
- The user id is any non-empty string. Node doesn't require a UUID, so Go doesn't either.

### Conversations

| Route | Rules (in order) |
|---|---|
| `POST /v1/conversations` | 1. `type` must be `dm` or `group`; otherwise invalid_input. 2. `participantIds` must be a non-empty array; otherwise invalid_input. 3. Every id must be a string of 36 chars or fewer; otherwise invalid_input. 4. Look up identities for `[me, ...participantIds]` via the user service. If the lookup fails: internal_server. 5. If my own identity is missing: invalid_input. If my username is blank after trim: user_setup_incomplete. 6. For each participant: identity missing → invalid_input; blank username → user_setup_incomplete. 7. **DM:** exactly 1 participant, else invalid_input. If that participant is me: cannot_message_self. If a DM with exactly these 2 participants exists, return it with **200**. 8. **Group:** at least 1 participant, else invalid_input. If participants + 1 > 50: too_many_participants. 9. Insert. The creator's role is `owner` for a group and `member` for a DM; everyone else is `member`. `name` = body name, or null if empty, for a group, and always null for a DM. Return **201** with the document |
| `GET /v1/conversations` | `page` = max(0, int or 0). `limit` = clamp(int or 20, 1, 50). Find by `participants.userId`, sort `updatedAt` desc, skip `page*limit`. Meta is `{page, page_size, total}` |
| `GET /v1/conversations/unread-counts` | For each of my conversations: count messages with `senderId != me`, `isDeleted == false`, and `_id > lastReadMessageId` when I have one. Only counts above 0 are included. Returns a `{conversationId: count}` map |
| `GET /v1/conversations/:id` | Invalid ObjectId or not found: conversation_not_found. Not a participant: not_participant |
| `PUT /v1/conversations/:id` | Not found (incl. bad id) → not a group (forbidden) → not a participant → role not owner/admin (insufficient_role). Then set `name` / `avatarUrl` when the key is present in the body |
| `DELETE /v1/conversations/:id` (leave) | Not found → not a participant. DM: delete the conversation and all its messages. Group: remove me. If nobody is left, delete the conversation and its messages. Otherwise, if no owner remains, promote the first admin, or the first remaining participant. Returns 200 with no data |
| `POST /v1/conversations/:id/participants` | 1. `userId` and `username` must both be truthy, else invalid_input (`username` is checked but then ignored). 2. Identity lookup: failure → internal_server; missing → invalid_input; blank username → user_setup_incomplete. 3. Not found → not a group (forbidden) → actor not a participant → actor role not owner/admin. 4. Already a participant: **200** with the unchanged document and no write. 5. Participants already ≥ 50: too_many_participants. 6. Append as `member` |
| `DELETE /v1/conversations/:id/participants/:userId` | Not found → not a group → actor not a participant → actor not owner/admin → target not a participant (not_participant) → target is the owner (insufficient_role). Then remove the target |

### DM messages

| Route | Rules (in order) |
|---|---|
| `GET …/:id/messages` | `limit` = clamp(int or 50, 1, 100). Not found → not a participant. If `before` is a valid ObjectId, filter `_id < before` (an invalid `before` is ignored). Sort `createdAt` desc, limit, then return in reverse (oldest first) |
| `POST …/:id/messages` | Handler: `text` must be a non-empty string, else invalid_input. If `type` is present, it must be `text`, `image` or `system`. Service: not found → not a participant. Blank text after trim, or `text.length > 2000` on the untrimmed text: invalid_input. `senderUsername` comes from the participant record. Insert `text` trimmed. `imageUrls` = body `imageUrls` only when the type is `image`, else `[]`. `replyTo` = ObjectId if it's a valid one, else null. `readBy` = `[{userId: me, readAt: now}]`. Then set `conversation.lastMessage = {_id, senderId, senderUsername, text: trimmed[:100], createdAt}`, which bumps `updatedAt`. Return **201** with the message plus `participantIds` |
| `PATCH …/:id/messages/:msgId` | Handler: `text` must be a non-empty string. Service: either id not a valid ObjectId → dm_message_not_found. Blank text after trim, or length > 2000 → invalid_input. Find by `{_id, conversationId}`: missing → dm_message_not_found; not the sender → forbidden; already deleted → dm_message_not_found. Set `text` trimmed |
| `DELETE …/:id/messages/:msgId` | Either id not a valid ObjectId, or not found → dm_message_not_found. Not the sender → forbidden. Set `isDeleted: true, text: ""` |
| `POST …/:id/read` | Not found → not a participant. If body `messageId` is a valid ObjectId, read up to it. Otherwise find the latest message by `createdAt`; if there is none, return 200 without writing. Set my participant's `lastReadMessageId`. Returns 200 with no data |

Not part of parity: membership checks don't apply to edit, delete or read beyond what's listed above. For example, edit and delete check the sender, not membership, and read doesn't check that `messageId` belongs to the conversation. That matches Node.

### Response envelope

- Use the shared `Response[T]` with chat templates for the 20000-range generic codes plus 50018–50026, copied from `types/api-response.ts` (codes, keys, statuses and messages).
- `RES_SUCC_CREATED` is status 201 with code 100000 and key `res_succ_ok`, as in Node.
- Go always sends `message`, while Node omits it on generic templates. Go also omits a `meta.total` of 0. The web already handles both (`lib/query/paginated.ts`).

### Document shape (Mongo)

| Collection | Fields (Mongoose defaults written explicitly by Go) |
|---|---|
| `conversations` | `_id` ObjectId, `type`, `name` (null), `avatarUrl` (null), `createdBy`, `participants[]` {`userId`, `username`, `profilePicture` (null), `role`, `joinedAt` Date, `lastReadMessageId` ObjectId or null, `isMuted` false} with no subdocument `_id`, `lastMessage` (null or {`_id`, `senderId`, `senderUsername`, `text`, `createdAt`}), `createdAt`, `updatedAt`, `__v` |
| `dmmessages` | `_id`, `conversationId` ObjectId, `senderId`, `senderUsername`, `type`, `text`, `imageUrls` ([]), `replyTo` (null), `isDeleted` (false), `readBy[]` {`userId`, `readAt`} with no subdocument `_id`, `createdAt`, `updatedAt`, `__v` |

- Every timestamp Go writes is `time.Now().UTC().Truncate(time.Millisecond)`. Mongo stores milliseconds, so this keeps the response equal to what was stored.
- Inserts set `__v: 0`. Writes that rewrite the `participants` array also `$inc` `__v`, as Mongoose versioning does, so Node's version-guarded saves keep working during the move.
- `updatedAt` changes only when a write actually changes something, like Mongoose's dirty tracking. For example, marking read at the same message or adding an existing participant doesn't bump it.
- Go uses targeted updates (`$set`, `$push`, `$pull`, positional `participants.$`) rather than Node's load–modify–save. The resulting documents are the same.
- Indexes are created at startup with the same keys as Mongoose:
  - `conversations`: `{participants.userId:1, updatedAt:-1}` and `{type:1, participants.userId:1}`
  - `dmmessages`: `{conversationId:1}` and `{conversationId:1, createdAt:-1}`
  - An index-creation error is logged and doesn't stop startup. An existing index with the same name but different options mustn't take the service down.

### JSON shape (responses and push payloads)

- Field names match the Mongoose `toObject()` output above, including `_id` and `__v`.
- ObjectIds are hex strings. Dates are `2006-01-02T15:04:05.000Z` (UTC, milliseconds), which is what `JSON.stringify(Date)` produces.
- Nullable fields are `null`, never omitted.
- Mongo documents (`domains`) and JSON (`dto`) use separate structs, with `dto.FromConversation` / `dto.FromDmMessage` converters. That keeps the date format in one place.

### Deliberate deviations (Node behaviour that's broken today)

Approved on 2026-10-06 ("fix that up, make your best change"). Items 4 and 5 were added during implementation.

1. **Deleting a DM message works in Go.** In Node it always fails: Mongoose's `required` validator rejects the empty `text` that delete sets, so the save throws and the client gets `404 res_err_route_not_found`. The web never calls delete, so nothing visible changes.
2. **Errors get real status codes.** Node turns every thrown error into `404 res_err_route_not_found`: malformed JSON, a Mongo failure, a schema `maxlength` violation. Go returns:
   - `400 res_err_invalid_payload` for malformed JSON
   - `400 res_err_invalid_input` for the schema limits Node enforced through Mongoose: `name` ≤ 100, `avatarUrl` ≤ 2048, `username` ≤ 50, `userId` ≤ 36
   - `500 res_err_database_issue` for a Mongo failure
3. **Your own message no longer bumps your unread count.** Today `dm:new_message` goes to the sender's tabs too, and a tab not viewing that conversation increments its unread badge for your own message. The server-side count already excludes your messages. This was only visible on `/messages`, because that's the only place the DM socket was mounted. Once the hook runs on every page it would show up everywhere, so the web skips the increment when `senderId` is you.
4. **Duplicate participant ids are dropped on create.** Node stored each repeat as another participant, and in a group it could add the creator a second time. Go de-duplicates the ids and leaves the creator out of a group's list (the creator is always added as owner). A DM created with `[other, other]` now succeeds instead of failing with invalid_input.
5. **Wrong JSON types get invalid_input.** Where Node let Mongoose coerce a value (a number as `name`, a single string as `imageUrls`) or crash on it, Go only accepts the documented type and otherwise returns `400 res_err_invalid_input`.
6. **A message in a conversation you don't have cached refreshes the conversation list.** This makes a DM that someone just started with you appear without a reload. Node's handler only updated conversations already in the cache.

### Known gaps carried over unchanged

- No presence snapshot: a user who comes online doesn't learn who is already online. Node never sent one, and the web's `setOnlineUsers` is unused.
- Presence fans out to contacts in your 100 most recently updated conversations, as Node's `getConversations(userId, 0, 100)` does.
- `dm:message_edited` / `dm:message_deleted` are still not emitted.

---

## Tasks

### Task 1: Rename Node chat to `chat-legacy`

- `git mv backend/chat backend/chat-legacy`
- `chat-legacy/src/index.ts` `CreateConsulRegistry`: set `serviceName: 'chat-legacy'`, `hostname: 'chat-legacy'` and `healthCheckURL: 'http://chat-legacy:7780/v1/health'`.
- `docker-compose-dev.yaml`: rename service `chat` to `chat-legacy`, with `container_name: letslive-chat-legacy`, `build.context: ./backend/chat-legacy/` and `ports: "7786:7780"`.
- `docker-compose.yaml`: the same, with `image: sen1or/letslive-chat-legacy:latest`.
- In both compose files, point the swagger volume at `./backend/chat-legacy/src/docs/openapi.yaml`.
- `.github/workflows/test.yml`: change the Node job's `cache-dependency-path` and `working-directory` to `./backend/chat-legacy`.
- `.github/workflows/build-and-publish-images.yml`: add `{path: ./backend/chat-legacy, name: letslive-chat-legacy, context: ./backend/chat-legacy}`. Task 2 repoints the existing `letslive-chat` entry.

### Task 2: Go module skeleton

- `backend/chat/go.mod` (`module sen1or/letslive/chat`, go 1.26.0) with dependencies on `sen1or/letslive/shared`, `mongo-driver/v2`, `nats.go`, `golang-jwt/jwt/v5`, `otelhttp` and `zap`. Add a `replace` for shared as `realtime` does, add `./chat` to `backend/go.work`, and run `go mod tidy`.
- `backend/chat/Dockerfile`: copy `realtime/Dockerfile` with `chat` swapped in.
- `config/config.go`:
  - `Service`, `Database{Host, Port, Name, Params}`, `NATS{URL}` and `Tracer`, plus the tracer interface methods.
  - `PostProcess` builds `mongodb://CHAT_DB_USER:CHAT_DB_PASSWORD@host:port/name?params`, URL-escaping the credentials, and fails if `nats.url` is empty.
- `cmd/main.go`, modelled on `realtime/cmd/main.go`:
  1. logger, Consul registry, config manager, discovery registration, OTel
  2. Mongo connect with retry, then ping
  3. `EnsureIndexes`
  4. NATS connect and presence subscription
  5. HTTP server
  6. graceful shutdown: server, deregister, OTel, unsubscribe and drain NATS, disconnect Mongo
- `api/server.go`:
  - `http.ServeMux` with the routes in Task 6, `GET /v1/health` (Mongo ping + NATS connected), and a catch-all `GET /` that returns route_not_found.
  - Wrap with `otelhttp` and the shared logging and request-ID middlewares, as `finance/api/http.go` does.
- CI:
  - Add `chat` to the Go test matrix in `test.yml`.
  - Point the `letslive-chat` image entry at `{path: ./backend/chat, context: ./backend}`.
- Compose:
  - Add a Go `chat` service to both files: dev builds `./backend/` with `chat/Dockerfile`; prod uses `image: sen1or/letslive-chat:latest`. Port `7780:7780`.
  - Env: `CONFIG_SERVER_PROFILE`, `CONFIG_SERVER_INTERVAL`, `REGISTRY_SERVICE_ADDRESS`, `CHAT_DB_USER`, `CHAT_DB_PASSWORD`.
  - `depends_on`: consul, nats and chat_db (healthy).
- `letslive-configs`: add `chat_service-dev.yml` / `chat_service-prod.yml` with:
  - service `chat`, hostname `chat`, port 7780
  - database `chat_db:27017/chat`, params `authSource=admin`
  - `nats://nats:4222`
  - the same tracer block as `realtime`

  I'll push these only when you say so.

### Task 3: Domain, DTO and response

- `domains/`: `Conversation`, `Participant`, `LastMessage`, `DmMessage`, `ReadReceipt` with `bson` tags matching the document table; enums `ConversationType`, `ParticipantRole`, `DmMessageType`; domain errors (`ErrConversationNotFound`, `ErrNotParticipant`, `ErrInsufficientRole`, `ErrDmMessageNotFound`, `ErrCannotMessageSelf`, `ErrTooManyParticipants`, `ErrUserSetupIncomplete`, `ErrForbidden`, `ErrInvalidInput`, `ErrDatabaseIssue`, `ErrUserService`).
- `dto/`: request structs, response structs with JSON tags, converters, and the date formatter.
- `response/`: type aliases to `shared/response`, all templates from Node's `api-response.ts`, and `FromError`, following `finance/response/from_error.go`. Map `ErrUserService` to `internal_server`, as Node does.

### Task 4: Repositories

`repositories/conversation.go` and `repositories/dm_message.go`. All use `context` timeouts from the request.

- Conversations:
  - `FindByID`; `FindDM(a, b)`, which is `type: dm`, `participants.userId $all [a,b]`, size 2
  - `ListForUser(userID, skip, limit)`, `CountForUser`, `ListAllForUser` (for unread counts)
  - `Insert`, `UpdateFields`
  - `AddParticipant` (`$push` + `$inc __v`); `SetParticipants` (`$set` of the array + `$inc __v`, used by remove/leave/owner transfer)
  - `SetLastMessage`, `SetLastRead` (positional, only if it changed), `Delete`
- Messages:
  - `Insert`, `FindPage`, `FindInConversation(id, convID)`, `UpdateText`, `SoftDelete`, `FindLatest`
  - `CountUnread(convID, afterID, excludeSender)`, `DeleteByConversation`
- `EnsureIndexes(ctx, db)`
- Mongo errors are wrapped as `ErrDatabaseIssue`. `mongo.ErrNoDocuments` becomes `nil, nil`, and the service decides which not-found error applies.

### Task 5: Services and user gateway

- `gateway/userservice`: `GetIdentities(ctx, ids)` dedupes the ids, does `POST http://<user>/v1/internal/users/batch` `{ids}` with a 3s timeout and the `X-Request-ID` header, and returns `map[id]Identity{Username, ProfilePicture *string}`. Any non-2xx or transport error returns `ErrUserService`.
- `services/conversation.go` and `services/dm_message.go`: each rule from the parity tables, in the same order, returning `(result, error)`.
- `SendMessage` returns the message and the participant ids. `MarkAsRead` returns whether it ran. `Typing` returns the sender's username and the other participants' ids.

### Task 6: Handlers and routes

Handlers in `handlers/conversation` and `handlers/dmmessage`, one file per route as in `finance`, using a `basehandler.WriteResponse` copy:

```
GET    /v1/conversations/unread-counts
GET    /v1/conversations
POST   /v1/conversations
GET    /v1/conversations/{id}
PUT    /v1/conversations/{id}
DELETE /v1/conversations/{id}
POST   /v1/conversations/{id}/participants
DELETE /v1/conversations/{id}/participants/{userId}
GET    /v1/conversations/{id}/messages
POST   /v1/conversations/{id}/messages
PATCH  /v1/conversations/{id}/messages/{msgId}
DELETE /v1/conversations/{id}/messages/{msgId}
POST   /v1/conversations/{id}/read
POST   /v1/conversations/{id}/typing          (new)
```

- Body fields that Node type-checks (`text` is a string, `participantIds` is an array of strings) are decoded into `json.RawMessage` / `any` where a type mismatch must produce `invalid_input` rather than `invalid_payload`. This keeps Node's error codes.
- `POST …/typing` takes `{"state": "start" | "stop"}`; any other value is invalid_input. Not found → not a participant. Returns 200 with no data.

### Task 7: Realtime publishing

Event names go in `chat/realtimeevents` as constants equal to the Node `DmServerEventType` values. Each push goes to `realtime.UserTopic(id)` for every recipient. A failure is logged with the request context and never changes the response.

| Trigger | Event | Recipients | `data` |
|---|---|---|---|
| `POST …/messages` succeeds | `dm:new_message` | all participants, sender included (multi-tab) | `{conversationId, message}` with the same message JSON as the REST response, minus `participantIds` |
| `POST …/typing` `start` / `stop` | `dm:user_typing` / `dm:user_stopped_typing` | other participants | `{conversationId, userId, username}` |
| `POST …/read` succeeds | `dm:read_receipt` | other participants | `{conversationId, userId, messageId?, readAt}`. `messageId` is omitted when the body had none, as in Node |

Today `POST …/read` over REST pushes nothing; only the socket's `dm:mark_read` did, and the web never sent it. Pushing on the REST call is new but harmless: the web's handler for `dm:read_receipt` is a no-op.

### Task 8: Presence subscriber

- `presence/subscriber.go` subscribes to `realtime.PresenceSubject` on the core NATS connection and decodes `realtime.PresenceEvent`; a bad payload is logged and dropped.
- It loads the user's 100 most recently updated conversations, collects every other participant id, and publishes `dm:user_online` or `dm:user_offline` with `data {userId}` to each contact.
- Each event gets its own 5s context. The NATS callback runs the work inline, which keeps events in order per subscription.

### Task 9: Kong

In `configs/kong.yml`:

- `Chat` (`chat.service.consul:7780`, now Go) keeps only the `DM_Conversations` route, JWT plugin unchanged.
- New service `Chat_Legacy` (`chat-legacy.service.consul:7780`, `path: /v1`, same timeouts) takes the `Chat_Messages` and `Chat_Commands` routes unchanged.
- `Chat_WS` changes its host to `chat-legacy.service.consul` and drops the `DM_Socket` (`/dm-ws`) route.

### Task 10: Web

- `lib/api/dm.ts`:
  - `SendDmMessage` drops the unused `senderUsername` from its body type.
  - Add `SendDmTyping(conversationId, state: "start" | "stop")`.
- `lib/query/dm-cache.ts`: `appendDmMessage` skips a message whose `_id` is already cached. A reconnect refetch racing a push could otherwise show it twice.
- New `hooks/use-dm-realtime.ts` replaces `use-dm-websocket.ts`:
  - It registers `onEvent` for each DM server event type and rebuilds `{type: frame.type, ...frame.data}` into the existing `DmWsServerEvent` shape.
  - It reuses the current switch, minus `SEND_FAILED`, with the typing timeouts and cleanup.
  - It skips `incrementDmUnread` when `message.senderId` is the current user.
  - On reconnect it invalidates the DM unread counts, the conversations list and the active conversation's messages.
- Mount `useDmRealtime(!!user)` in `app/(main)/_components/header/messages-icon.tsx`, the same way the bell mounts `useNotificationRealtime`, so unread counts and presence stay live on every page.
- `hooks/queries/use-dm-unread-counts.ts`: remove `refetchInterval` / `refetchIntervalInBackground`.
- `app/(main)/messages/[conversationId]/page.tsx`:
  - Send with `SendDmMessage` (type `image` when there are `imageUrls`). On failure, toast `t(key)` with the `"Failed to send message"` fallback, as the `SEND_FAILED` handler did. The pushed `dm:new_message` adds the message to the thread, so the page doesn't append the REST result itself.
  - Typing start and stop call `SendDmTyping` and ignore errors. The input already sends `start` once per burst and `stop` after 2s idle, so no extra throttling is needed.
- `app/(main)/messages/layout.tsx`: remove `DmWebSocketProvider`.
- Delete `contexts/dm-websocket-context.tsx` and `hooks/use-dm-websocket.ts`. Remove `DM_WS_SERVER_URL` from `global.ts`, and remove `DmClientEventType`, the client event types and `DmWsSendFailed` from `types/dm.ts`. `WS_SERVER_URL` stays until Phase 3.

### Task 11: Verify

- `gofmt -l`, `go build ./...` and `go vet ./...` for `chat`, `shared` and `realtime`.
- `docker build` for `chat` (Go) and `chat-legacy`. Parse both compose files with `docker compose config`.
- `web`: `npx tsc --noEmit` and `npm run lint`.
- `chat-legacy`: `npm run build`, to confirm the rename broke nothing.
- Parity review: a fresh reviewer reads every Go service function next to its Node counterpart, using the tables above as the checklist, and checks the document shape against `models/*.ts`.

### Running-stack checks (for you)

1. Go `chat` and `chat-legacy` both register in Consul and report healthy.
2. Against existing data written by Node: the conversation list, a conversation's messages and unread counts look the same as before.
3. Create a DM and a group from the web. Compare the new documents in `chat_db` with older Node-written ones (`db.conversations.find().sort({_id:-1}).limit(2)`).
4. Two browsers, two users:
   - a message appears instantly for both
   - the typing indicator shows and clears
   - the unread badge goes up in the header on a non-messages page and clears when the conversation is opened
   - presence goes online and offline (offline about 5s after the last tab closes)
5. Two tabs as the same user: both get new messages, and the sender's own badge doesn't go up.
6. Group actions: rename, add and remove a participant, leave as owner (ownership transfers), and the last member leaving deletes the group.
7. Live chat (`/ws`, `/messages`, `/chat-commands`) still works through `chat-legacy`.
8. Rollback drill (optional): point Kong `Chat` at `chat-legacy.service.consul`. DM REST still works.

---

## Files touched

| Area | Files |
|---|---|
| New | `backend/chat/**` (Go), `web/hooks/use-dm-realtime.ts`, `letslive-configs/chat_service-{dev,prod}.yml` |
| Moved | `backend/chat/**` → `backend/chat-legacy/**` |
| Edited | `backend/chat-legacy/src/index.ts`, `backend/go.work`, `docker-compose.yaml`, `docker-compose-dev.yaml`, `configs/kong.yml`, `.github/workflows/test.yml`, `.github/workflows/build-and-publish-images.yml`, `web/lib/api/dm.ts`, `web/lib/query/dm-cache.ts`, `web/hooks/queries/use-dm-unread-counts.ts`, `web/app/(main)/_components/header/messages-icon.tsx`, `web/app/(main)/messages/layout.tsx`, `web/app/(main)/messages/[conversationId]/page.tsx`, `web/global.ts`, `web/types/dm.ts` |
| Deleted | `web/contexts/dm-websocket-context.tsx`, `web/hooks/use-dm-websocket.ts` |

## Out of scope

- Live chat and chat commands in Go (Phase 3).
- Deleting Node's `/dm-ws` code, Redis and `Chat_WS` (Phase 4). The Node DM socket code stays in `chat-legacy` with no route to it.
- OpenAPI for the Go chat service (Phase 4).
