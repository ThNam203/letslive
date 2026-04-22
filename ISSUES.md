# Issues — letslive

_Last updated: 2026-04-17_

---

## Security Issues

### 🔴 CRITICAL

**S1. JWT signature NOT verified in any service**
All Go services call `jwt.ParseUnverified()`; the chat service calls `jwt.decode()`. Anyone who can reach a service directly — bypassing Kong — can forge any user's identity.
Files: [user/handlers/utils/cookie.go:28](user/handlers/utils/cookie.go#L28), [livestream/handlers/utils/cookie.go:28](livestream/handlers/utils/cookie.go#L28), [vod/handlers/utils/cookie.go:28](vod/handlers/utils/cookie.go#L28), [backend/chat/src/middlewares/auth.ts:14](backend/chat/src/middlewares/auth.ts#L14)

**S2. Refresh token NOT revoked on logout**
`LogOutHandler` clears the cookie but never calls `RevokeTokenByValue()` or `RevokeAllTokensOfUser()`. A stolen refresh token stays valid after logout indefinitely.
File: [backend/auth/handlers/auth.go:269-272](backend/auth/handlers/auth.go#L269-L272)
Methods exist but unused: `RevokeTokenByValue()`, `RevokeAllTokensOfUser()`

**S3. Revoked tokens accepted in RefreshToken flow**
`RefreshToken()` validates the JWT signature and expiry but never queries the `revoked_at` field in the `refresh_tokens` table. Explicitly revoked tokens can still mint new access tokens.
File: [backend/auth/services/jwt.go:51-85](backend/auth/services/jwt.go#L51-L85)

---

### 🟠 HIGH

**S4. Unauthenticated file upload**
`POST /v1/upload-file` has no auth middleware, no MIME/extension validation, and the MinIO bucket is public-read. Enables malware hosting and possible RCE.
File: [user/api/http.go:55](user/api/http.go#L55)

**S5. CORS wildcard with credentials**
Kong is configured with `origins: ["*"]` and `credentials: true`. Any website can issue authenticated requests and read responses.
File: [configs/kong.yml:431](configs/kong.yml#L431)

**S6. No TLS on backend services**
The `useTLS` path hard-errors. JWT cookies and payloads travel in plaintext between Kong and all services (and to clients if Kong is not fronted by TLS).
File: [auth/api/http.go:58](auth/api/http.go#L58)

**S7. Broken rate limiting — OTP endpoint**
OTP rate limit is set to 123/min with a `#TODO: change to one` comment. No brute-force protection on login or signup.
File: [configs/kong.yml:76-80](configs/kong.yml#L76-L80)

**S8. Hardcoded secrets committed to version control**
The `.env` file is tracked in the repo and contains live secrets:
- `ACCESS_TOKEN_SECRET=access_token_secret`
- `REFRESH_TOKEN_SECRET=refresh_token_secret`
- `GMAIL_APP_PASSWORD="0000 0000 0000 0000"`
- `CONFIG_SERVER_GIT_PASSWORD=` ← live GitHub PAT
- All service database credentials
File: [.env](.env)

**S9. Weak JWT signing secret in Kong**
Kong's JWT consumer secret is the literal string `"access_token_secret"`. A comment confirms it: `## note to myself: keep it access_token_secret, I use it to replace with real secret`. Any attacker can forge valid access tokens.
File: [configs/kong.yml:435](configs/kong.yml#L435)

---

### 🟡 MEDIUM

**S10. Finance service routed but unimplemented**
Kong routes `/finance` to a service that only exposes `/health`. When real endpoints land they must enforce per-user authz, atomic balance updates, and rejection of negative amounts to prevent double-spend.
File: [finance/api/http.go](finance/api/http.go)

**S11. Chat conversation updates lack role checks**
Any participant — not just the owner — can rename or modify a group conversation.
File: [backend/chat/src/services/conversationService.ts:143](backend/chat/src/services/conversationService.ts#L143)

**S12. Stored XSS via unvalidated chat fields**
`name`, `displayName`, and `profilePicture` URL fields in chat handlers are stored without sanitization. Frontends that render them as HTML are vulnerable to stored XSS.

**S13. Participant identity spoofing on conversation creation**
When creating a conversation, `displayName`, `profilePicture`, and `username` values are taken directly from the client request body with no validation that they match actual user records. A malicious client can impersonate other participants.
File: [backend/chat/src/handlers/conversationHandler.ts:44-49](backend/chat/src/handlers/conversationHandler.ts#L44-L49)

**S14. Race condition in `leaveConversation`**
The function reads conversation state, mutates it in memory, then writes back. Concurrent leave/remove operations (e.g., two admins acting simultaneously) can corrupt the owner-transfer logic and leave conversations in an inconsistent state.
File: [backend/chat/src/services/conversationService.ts:258-298](backend/chat/src/services/conversationService.ts#L258-L298)

**S15. Swagger UI publicly exposed**
The Swagger UI is exposed on port 11111 with no authentication, leaking the full API surface area.
File: [docker-compose.yaml:191-203](docker-compose.yaml#L191-L203)

**S16. CSRF — cookie-based JWT with no CSRF token**
Combined with the CORS wildcard (S5), cookie-based auth with no CSRF token allows cross-site request forgery. Fix: add `SameSite=Strict` or per-request CSRF tokens. Fixing CORS alone (S5) mitigates most of this.

---

### 🟢 LOW

**S17. Missing security headers at the gateway**
HSTS, `X-Frame-Options`, `X-Content-Type-Options`, and CSP are absent.

**S18. Default admin credentials**
Grafana and MinIO use `admin:admin` / `postgres:postgres`. Acceptable for dev; must not ship to prod.
File: [docker-compose.yaml:62-63](docker-compose.yaml#L62-L63)

**S19. Weak input validation on livestream fields**
`title` and `description` are only checked for max length. No whitespace-only check, no HTML filtering.
File: [backend/user/handlers/livestream_information/update_private.go:52-63](backend/user/handlers/livestream_information/update_private.go#L52-L63)

---

## Logic / Consistency Issues

### 🟠 HIGH

**L1. WebSocket validation fails silently**
When a DM WebSocket event fails validation (missing `conversationId`, empty text, text > 2000 chars), the handler silently `return`s with no error event sent to the client. The REST handler returns proper error responses. Users have no feedback when a message is dropped.
File: [backend/chat/src/dmServer.ts:93-99](backend/chat/src/dmServer.ts#L93-L99)

**L2. `CreateConversationRequest` type is incomplete**
The frontend sends `participantUsernames`, `participantDisplayNames`, `participantProfilePictures`, `creatorUsername`, `creatorDisplayName`, and `creatorProfilePicture` — none of which appear in the backend type definition. The handler casts to the type then accesses `req.body` directly for the extra fields, bypassing type safety entirely.
Backend type: [backend/chat/src/types/conversation.ts:18-22](backend/chat/src/types/conversation.ts#L18-L22)
Frontend call: [web/lib/api/dm.ts:27-43](web/lib/api/dm.ts#L27-L43)

**L3. WebSocket event types missing required fields**
- `DmSendMessageEvent` is missing `senderUsername` — the frontend sends it, the backend handler reads it from raw data, but it is absent from the type definition.
- `DmTypingEvent` is missing `username` — same pattern.
Backend types: [backend/chat/src/types/dm-event.ts:11-18](backend/chat/src/types/dm-event.ts#L11-L18)
Frontend types: [web/types/dm.ts:90-98](web/types/dm.ts#L90-L98)

---

### 🟡 MEDIUM

**L4. Inconsistent field name for message type across protocols**
The REST API uses `type?: DmMessageType`; the WebSocket type uses `messageType: DmMessageType`. Developers must track which name applies to which transport.
REST: [web/lib/api/dm.ts:110](web/lib/api/dm.ts#L110)
WebSocket: [web/types/dm.ts:94](web/types/dm.ts#L94)

**L5. `GetMessages` silently swallows errors**
The function catches all errors and returns `{ messages: [] }` rather than propagating the error. Callers cannot distinguish a successful empty result from a network/server failure. All other API functions in the same file return `ApiResponse<T>`.
File: [web/lib/api/chat.ts:5-16](web/lib/api/chat.ts#L5-L16)

---

### 🟢 LOW

**L6. Timestamp format inconsistency across message types**
`ReceivedMessage` (chat) uses `timestamp: number`; DM messages use `createdAt: string`. The types represent different systems, but the inconsistency makes shared utilities error-prone.
File: [web/types/message.ts:15](web/types/message.ts#L15)

**L7. `uuid.FromStringOrNil` silently returns nil UUID on failure**
Instead of returning an error, the conversion silently yields a nil UUID, masking upstream bugs.
File: [backend/user/handlers/user/update_current_user_private.go:45](backend/user/handlers/user/update_current_user_private.go#L45)

**L8. No DB-level uniqueness for active livestream per user**
The app reads only one active livestream (`GET /v1/livestreams?userId=...`) and now picks the latest one in query order, but the database schema still allows multiple active rows for the same user. This can create nondeterministic behavior across other queries and background jobs.
Suggested future fix: add a partial unique index on `livestreams(user_id)` where `vod_id IS NULL AND ended_at IS NULL` after data cleanup.
File: [backend/livestream/migrations/0001_init_tables.sql](backend/livestream/migrations/0001_init_tables.sql)

---

## Recommended Fix Order

1. **S8** — Rotate all leaked secrets immediately; remove `.env` from git history
2. **S1** — Verify JWT signatures in all services
3. **S2 + S3** — Enforce token revocation on logout and in the refresh flow
4. **S9** — Replace Kong JWT secret with a random value via env var
5. **S4** — Add auth middleware and MIME validation to the upload endpoint
6. **S5 + S16** — Fix CORS; add `SameSite=Strict` to cookies
7. **S7** — Set OTP rate limit to 1/min
8. **S6** — Enable TLS between Kong and services
9. **L1** — Send error events to WebSocket clients on validation failure
10. **L2 + L3** — Align backend type definitions with actual runtime payloads
