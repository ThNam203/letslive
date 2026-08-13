# Issues — letslive

_Last updated: 2026-07-14_

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

**S10. ~~Finance service routed but unimplemented~~ — SUPERSEDED**
Finance service is now implemented (branch `feat/finance-service`). Per-user authz, atomic balance updates, and negative-amount rejection are in place (double-entry ledger with DB triggers). Remaining finance findings tracked in the **Finance Service Issues** section below (F1–F15).

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

## Finance Service Issues

_Audit of `backend/finance` on branch `feat/finance-service`, 2026-07-14. Ledger core verified sound: zero-sum DB trigger, ledger immutability triggers, `FOR UPDATE` completion idempotency, wallet overdraft rejection. Build, vet, and tests pass._

### 🔴 HIGH

**F1. Refund compensation dies with the request context → stranded funds**
`Purchase()` debits the wallet, then calls the user service with the request-scoped `ctx`. If the client disconnects (or the server's 10s `WriteTimeout` fires) after the debit, `CreateGift`/`AddInventory` fails with `context canceled` — and the compensating `refund()` then runs its DB writes on the same canceled context, so it fails too. Money leaves the wallet, no item is granted, no refund lands; the only trace is a CRITICAL log requiring manual reconciliation.
Fix: run compensation (and arguably the post-debit grant call) on `context.WithoutCancel(ctx)`. Same for `failTransaction` in the deposit service.
Files: [backend/finance/services/purchase/purchase.go:103-149](backend/finance/services/purchase/purchase.go#L103-L149), [backend/finance/services/deposit/deposit.go:48](backend/finance/services/deposit/deposit.go#L48)

**F2. `http.DefaultClient` with no timeout in the user-service gateway**
Purchase holds committed ledger state while the inter-service call hangs; only the server write timeout eventually kills the connection — which then triggers F1. Give the gateway a dedicated `http.Client{Timeout: ~5s}`.
Files: [backend/finance/gateway/userservice/http/http.go:74](backend/finance/gateway/userservice/http/http.go#L74), [backend/finance/gateway/userservice/http/http.go:119](backend/finance/gateway/userservice/http/http.go#L119)

**F3. Stripe `checkout.session.completed` credits wallet without checking `payment_status`**
For async payment methods (bank debits, etc.), `completed` fires with `payment_status=unpaid`; the code credits the wallet immediately and marks the payment completed. A later `async_payment_failed` hits the terminal-state guard and is silently dropped — the user keeps the balance for a failed charge. `checkout.session.async_payment_succeeded` is also unhandled. Card-only checkout masks this today.
Fix: only credit when `session.payment_status == "paid"`; fulfill async payments on `async_payment_succeeded`.
Files: [backend/finance/gateway/payment/stripe/stripe.go:86-91](backend/finance/gateway/payment/stripe/stripe.go#L86-L91), [backend/finance/services/deposit/handle_webhook.go:46](backend/finance/services/deposit/handle_webhook.go#L46)

### 🟠 MEDIUM

**F4. Payment status has no transition guard (transactions do; payments don't)**
The terminal-state check in the webhook handler is check-then-act. A concurrent or late `failed` webhook after `completed` processing overwrites the payment row to `failed` while the wallet stays credited (the ledger itself is protected by the DB trigger; the payment row now lies).
Fix: guarded update — `UPDATE payments SET status=$1 WHERE id=$2 AND status IN ('created','processing')`.
Files: [backend/finance/repositories/payment/update_status.go:13](backend/finance/repositories/payment/update_status.go#L13), [backend/finance/services/deposit/handle_webhook.go:46-56](backend/finance/services/deposit/handle_webhook.go#L46-L56)

**F5. Global `stripego.Key` written on every checkout call**
Unsynchronized write to a package-level global per request — a data race under `-race`, and it breaks the moment a second Stripe key exists. Set once at init, or use a `client.API` instance.
File: [backend/finance/gateway/payment/stripe/stripe.go:42](backend/finance/gateway/payment/stripe/stripe.go#L42)

**F6. Missing index on `transactions(actor_id)`**
`ListByActor` runs `count(*)` and a paged list by `actor_id` — sequential scans on the busiest table. Add an index in the next migration.
File: [backend/finance/repositories/transaction/list_by_actor.go:14-32](backend/finance/repositories/transaction/list_by_actor.go#L14-L32)

**F7. No client idempotency key on purchase/deposit**
`transactions.reference` is documented as an idempotency key but is always a server-generated UUID — a double-click means a double purchase. Accept a client-supplied idempotency key in the DTOs and dedupe on the existing `reference` unique constraint.
Files: [backend/finance/migrations/0001_init_tables.sql:30](backend/finance/migrations/0001_init_tables.sql#L30), [backend/finance/services/purchase/purchase.go:88](backend/finance/services/purchase/purchase.go#L88), [backend/finance/services/deposit/initiate.go:78-88](backend/finance/services/deposit/initiate.go#L78-L88)

### 🟡 LOW

**F8. Ownership-check error mismatch (500 instead of 404)**
Wrong-owner transaction lookup returns `RES_ERR_TRANSACTION_FAILED` (HTTP 500); the payment service correctly returns 404 for the same case. Inconsistent and the wrong status code.
Files: [backend/finance/services/transaction/get_for_actor.go:16-23](backend/finance/services/transaction/get_for_actor.go#L16-L23), [backend/finance/services/payment/get_for_actor.go:21-28](backend/finance/services/payment/get_for_actor.go#L21-L28)

**F9. Unbounded webhook request body**
`io.ReadAll(r.Body)` with no `http.MaxBytesReader` on the public webhook endpoint. Cap it (e.g. 64 KB).
File: [backend/finance/handlers/deposit/handle_webhook_public.go:19](backend/finance/handlers/deposit/handle_webhook_public.go#L19)

**F10. Mock provider dead-ends in dev**
The deposit DTO allows `provider: mock` and the mock gateway registers under the dev profile, but the only webhook route is `/v1/deposits/webhook/stripe` — mock deposits can never complete, even in dev.
Files: [backend/finance/dto/deposit_request.go:4](backend/finance/dto/deposit_request.go#L4), [backend/finance/api/http.go:88](backend/finance/api/http.go#L88)

**F11. Unknown `provider_ref` returns 404 to Stripe → retry storm**
Webhook events for sessions created by another environment sharing the endpoint get 404, so Stripe retries for days. Consider returning 200 for unknown refs after signature verification.
Files: [backend/finance/services/deposit/handle_webhook.go:40-43](backend/finance/services/deposit/handle_webhook.go#L40-L43)

**F12. N+1 entries query in transaction listing**
`ListForActor` fetches ledger entries per transaction in a loop. Bounded by the max page size of 20 — acceptable for now, note only.
File: [backend/finance/services/transaction/list_for_actor.go:32-42](backend/finance/services/transaction/list_for_actor.go#L32-L42)

**F13. Implicit 1:1 platform-currency↔fiat peg**
Deposit amounts in platform-currency minor units are passed raw as Stripe `UnitAmount` in `fiatCurrencyCode`. Holds only while precision matches and the rate is 1:1. Make the rate explicit in config before it silently misprices.
File: [backend/finance/gateway/payment/stripe/stripe.go:51-55](backend/finance/gateway/payment/stripe/stripe.go#L51-L55)

**F14. Unauthenticated internal endpoints mint gifts/inventory**
`/v1/internal/gifts/create` and `/v1/internal/inventory/add` on the user service have no auth; they are safe only because Kong does not route them (extends S1's trust-the-network posture to money-adjacent writes). An internal API key header would harden this.
Files: [backend/user/api/http.go:88](backend/user/api/http.go#L88), [backend/user/api/http.go:94](backend/user/api/http.go#L94)

**F15. Minor cleanups**
- Dead `pgx.ErrNoRows` check on `Query` error (never returned there): [backend/finance/repositories/shop_item/list.go:23](backend/finance/repositories/shop_item/list.go#L23)
- Self-gifting allowed (no `actor == recipient` check in purchase) — confirm intended: [backend/finance/services/purchase/purchase.go:107](backend/finance/services/purchase/purchase.go#L107)
- No validation that `deposit.minAmount <= maxAmount` in config: [backend/finance/config/config.go:33-36](backend/finance/config/config.go#L33-L36)

---

## Web — TanStack Query Migration (Pending Manual Verification)

_Migration landed on `feat/finance-service`, 2026-08-13. Verified via `tsc --noEmit`, `eslint`, `next build`, and a dev-server smoke check (4 pages returned 200 with real HTML, no crash/error boundary). No live backend was running in the environment this was built in, so none of the authenticated data flows below were exercised against real data — needs a manual pass against a real backend before being trusted._

### 🟠 Needs manual click-through before trust

**W1. DM WebSocket ↔ query-cache sync (highest risk, do this one first)**
`use-dm-websocket.ts` now writes live events (new message, edit, delete, unread increment, conversation update) via `queryClient.setQueryData` instead of the old Zustand setters. This is the one part of the migration that can't be verified by build/typecheck alone — needs two logged-in browser sessions messaging each other, checking: live message delivery + correct chronological order, typing indicators, presence (online/offline), unread badge increments on the non-active conversation only, and reconnect behavior after a dropped socket.
Files: [web/hooks/use-dm-websocket.ts](web/hooks/use-dm-websocket.ts), [web/lib/query/dm-cache.ts](web/lib/query/dm-cache.ts)

**W2. Money-adjacent mutation coordination**
Shop purchase and gift-send now invalidate wallet balance + inventory together (previously separate manual refetch callbacks, easy to forget one). Verify: buy an item → balance decreases and item appears in inventory; send a gift to another user → sender balance decreases, sender inventory unchanged, recipient receives the item.
Files: [web/app/[lng]/(main)/shop/page.tsx](web/app/[lng]/(main)/shop/page.tsx), [web/app/[lng]/(main)/users/[userId]/gift-modal.tsx](web/app/[lng]/(main)/users/[userId]/gift-modal.tsx), [web/app/[lng]/(main)/wallet/inventory/send-gift-dialog.tsx](web/app/[lng]/(main)/wallet/inventory/send-gift-dialog.tsx)

**W3. Notification + DM polling replacement**
Manual `setInterval`/`visibilitychange` polling replaced with `useQuery({ refetchInterval, refetchOnWindowFocus })`. Verify unread counts still update in the background within ~30s and immediately on tab focus, for both the notification bell and the messages icon.
Files: [web/hooks/queries/use-notifications.ts](web/hooks/queries/use-notifications.ts), [web/hooks/queries/use-dm-unread-counts.ts](web/hooks/queries/use-dm-unread-counts.ts)

**W4. Pagination correctness**
Conversations, notifications, wallet transactions, and VOD comments moved to `useInfiniteQuery`. Verify "load more" actually appends (not replaces) and correctly stops at the end of each list — most likely place for an off-by-one if it breaks.
Files: [web/hooks/queries/use-conversations.ts](web/hooks/queries/use-conversations.ts), [web/hooks/queries/use-notifications.ts](web/hooks/queries/use-notifications.ts), [web/hooks/queries/use-transactions.ts](web/hooks/queries/use-transactions.ts), [web/hooks/queries/use-vod-comments.ts](web/hooks/queries/use-vod-comments.ts)

### 🟢 Lower risk, spot-check

**W5. Settings mutations** — profile (username/bio/pictures), stream info, password change, chat-commands CRUD. Straightforward `useMutation` wraps of existing logic; spot-check each save button once.
**W6. VOD comment create/delete/like** — optimistic local state preserved as-is, only the API-call plumbing changed.

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
9. **F1 + F2** — Compensation on `context.WithoutCancel`; HTTP client timeout in finance→user gateway (before merging `feat/finance-service`)
10. **F3 + F4** — Check Stripe `payment_status` before crediting; guarded payment status transitions
11. **L1** — Send error events to WebSocket clients on validation failure
12. **L2 + L3** — Align backend type definitions with actual runtime payloads
