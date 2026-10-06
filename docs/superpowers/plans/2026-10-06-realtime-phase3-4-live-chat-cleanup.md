# Realtime Phases 3–4: Live Chat in Go, Delete Node Chat

**Date:** 2026-10-06
**Author:** ThNam203
**Spec:** `docs/superpowers/specs/2026-10-06-realtime-gateway-design.md` (Phase 3, Phase 4)
**Builds on:** Phase 2 plan, `2026-10-06-realtime-phase2-go-chat-dms.md`. Same branch and PR (#367).

---

## Goal

- **Phase 3:** the Go `chat` service takes over live chat history, live chat send and chat commands. The web live chat panel moves onto the realtime socket.
- **Phase 4:** nothing routes to Node any more, so delete `backend/chat-legacy`, its Redis, the old Kong services and the old web socket code.

## Shape of the code

The Go service keeps the Phase 2 layout. Each feature is one vertical slice, so the files stay small:

| Feature | domains | repositories | services | handlers |
|---|---|---|---|---|
| Live chat | `live_message.go` | `live_message.go` | `live_chat.go` | `handlers/livechat/` |
| Chat commands | `chat_command.go` | `chat_command.go` | `chat_command.go` | `handlers/chatcommand/` |

These slices share what Phase 2 already has: `jsutil`, `handlers/utils`, `response`, `gateway/userservice`, and `events.Notifier`. `Notifier` gains a `ToRoom` method next to `ToUsers`.

## Phase 3 rules (Node → Go)

### Live chat

| Route | Rules |
|---|---|
| `GET /v1/messages?roomId=` (public) | An empty `roomId` or one longer than 36: `404 room_not_found`. Returns the latest 50 by `timestamp`, oldest first, as Mongoose docs (`_id`, `roomId`, `userId`, `username`, `text`, `profilePicture`, `timestamp`, `__v`) |
| `POST /v1/messages` `{roomId, text}` (JWT) | New. `roomId` must be a valid room topic id, else `room_not_found`. `text` must be a non-blank string of 500 or fewer JS chars, else `invalid_input`. `userId` comes from the cookie. `username` and `profilePicture` come from the user service; a blank username gives `user_setup_incomplete`. Save, then push `chat.message` to `room:<roomId>`. Returns 201 with the saved doc |

- The push payload is today's `ReceivedMessage`: `{id, userId, username, text, profilePicture, timestamp}`, with `timestamp` in epoch milliseconds as before.
- Join/leave lines come from the gateway's existing `member.joined` / `member.left` (`{userId}`). The web looks up the name and avatar with `usePublicUser`.

**Fixed over Node:**
1. **Impersonation:** the sender's name came from the client payload, so anyone could post as anyone. It now comes from the cookie identity and the user service.
2. **Empty messages:** Node broadcast an empty message, then failed to save it. Go rejects it.
3. **Anonymous viewers:** they now receive live chat. Before, they never sent `JOIN`, so they got nothing.

### Chat commands

Same routes, rules and order as `chatCommandService.ts`:
- name: trim + lowercase, must match `^[a-z0-9_-]{1,32}$`
- response: 1–500 JS chars
- description: 120 chars or fewer, default `""`
- scope: `user` or `channel`
- at most 50 per owner per scope
- a duplicate `(scope, ownerId, name)` gives `invalid_input`
- update: not found gives `invalid_input`; someone else's command gives `forbidden`
- delete: not found returns OK; someone else's command gives `forbidden`

Responses keep the Node shape: `{id, scope, ownerId, name, response, description, createdAt}`, and `/mine` returns `{user: [...], channel: [...]}` sorted by name.

**Fixed over Node:**
1. **Forged cookies:** the `/chat-commands` Kong route had no JWT plugin, and Node only decoded the cookie, so a forged one could create or delete commands as anyone. Kong now checks the JWT on everything except the public `GET /chat-commands?roomId=`.
2. **Malformed ids:** they returned `500 database_query`. They now count as not found.

### Kong

The `Chat` service gets two public routes, using the regex + method pattern that `Finance_Webhook_Route` already uses:
- `Chat_Public`: `GET ~/messages$` and `GET ~/chat-commands$`
- `Chat_Private` (JWT): `/messages`, `/chat-commands` and `/conversations`

### Web

- `RealtimeProvider` gains `subscribe(topic)`, as the spec planned. It ref-counts topics, sends `subscribe` / `unsubscribe`, and re-subscribes after a reconnect.
- Chat panel:
  - subscribes to `room:<roomId>`
  - appends `chat.message` lines for that topic
  - shows `member.joined` / `member.left` as join/leave lines
  - sends with `POST /messages`
  - on reconnect, refetches the backlog and drops the live lines
- Removed: `SendMessage`, the JOIN/LEAVE socket frames, `CHAT_MESSAGE_TYPE` and `WS_SERVER_URL`.

## Phase 4 deletions

- `backend/chat-legacy/**`
- compose (both files): `chat-legacy`, `chat_pubsub` and the `chat_pubsub_data` volume
- Kong: `Chat_Legacy`, `Chat_WS`
- CI: the Node test job and the `letslive-chat-legacy` image
- Docs:
  - `backend/chat/docs/openapi.yaml` (Go REST API) replaces the Node spec in swagger
  - new `backend/realtime/docs/openapi.yaml` (socket protocol)
  - `docs/CHAT_COMMANDS.md` and `docs/SYSTEM_DESIGN_QA.md` updated where they describe the old Node / Redis chat

## Verification

- Static: `go build` / `vet` / `gofmt`, the Docker image, `docker compose config`, and web `tsc` / lint / Prettier.
- Running stack: extend the Phase 2 end-to-end script with:
  - live chat history and send
  - push to a logged-in subscriber and an anonymous one
  - join/leave events
  - the impersonation fix
  - every chat command rule
  - Kong rejecting a forged cookie on chat-command writes

  Re-run the Phase 1–2 checks too.
