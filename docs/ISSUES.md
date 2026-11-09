# Issues — letslive

_Last updated: 2026-11-09_

---

## Security Issues

### 🔴 CRITICAL

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
The gateway's CORS policy allows any origin (`.*`) with `allow_credentials: true`, so any website can issue authenticated requests and read responses.
File: [configs/envoy/envoy.yaml](configs/envoy/envoy.yaml) (`envoy.filters.http.cors` policy on the virtual host)

**S6. No TLS on backend services**
The `useTLS` path hard-errors. JWT cookies and payloads travel in plaintext between the gateway and all services (and to clients if the gateway is not fronted by TLS).
File: [auth/api/http.go:58](auth/api/http.go#L58)

**S7. Broken rate limiting — OTP endpoint**
OTP rate limit is set to 123/min with a `#TODO: change to one` comment. No brute-force protection on login or signup.
File: [configs/envoy/envoy.yaml](configs/envoy/envoy.yaml) (the `/auth/verify-email` route, 123 tokens per 60s per client)

**S8. Hardcoded secrets committed to version control**
The `.env` file is tracked in the repo and contains live secrets:
- `ACCESS_TOKEN_SECRET=access_token_secret`
- `REFRESH_TOKEN_SECRET=refresh_token_secret`
- `GMAIL_APP_PASSWORD="0000 0000 0000 0000"`
- `CONFIG_SERVER_GIT_PASSWORD=` ← live GitHub PAT
- All service database credentials
File: [.env](.env)

---

### 🟡 MEDIUM

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

**S20. DM `imageUrls` are not validated server-side**
_Added 2026-11-05 from the Flutter mobile app final review._ Any string is stored as a picture URL and shown as a link and an image by web (`message-bubble.tsx`) and mobile. Mobile opens only `http`/`https` links, but the server should not store the rest.
Fix: accept only `https` URLs on the upload (MinIO) host, with a length cap.
File: [backend/chat/handlers/dmmessage/send_message_private.go:53-56](backend/chat/handlers/dmmessage/send_message_private.go#L53-L56)

**S21. `PATCH /auth/password` does not revoke refresh tokens**
_Added 2026-11-05 from the Flutter mobile app final review._ After a password change, other sessions keep refreshing. Related to S2/S3.
Fix: call `RevokeAllTokensOfUser` on success.
File: [backend/auth/services/auth.go:126](backend/auth/services/auth.go#L126)

**S27. VOD comments ignore VOD visibility**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ Listing, replying to and posting comments only check that the VOD exists, so anyone with a VOD id can read and write comments on a private or not-yet-ready VOD, and learn that it exists. `GetVODById`, view counting and reactions use `domains.VOD.VisibleTo`; comments do not. Older than this branch.
Fix: check `vod.VisibleTo(viewer)` in comment list, replies, create and like (the public list handlers need the optional cookie viewer, like `GetVODByIdPublicHandler`).
Files: [backend/vod/services/vod_comment/create.go](backend/vod/services/vod_comment/create.go), [backend/vod/services/vod_comment/get_comments.go](backend/vod/services/vod_comment/get_comments.go), [backend/vod/services/vod_comment/get_replies.go](backend/vod/services/vod_comment/get_replies.go)

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

**S22. Envoy does not normalize paths**
_Added 2026-11-05 from the Flutter mobile app final review._ The HTTP connection manager has no `normalize_path` or `merge_slashes`. Internal routes (`/v1/internal/*`) rely on Go's ServeMux redirecting unclean paths.
Fix: enable both, for defense in depth.
File: [configs/envoy/envoy.yaml](configs/envoy/envoy.yaml)

**S23. Turnstile verification HTTP call has no timeout**
_Added 2026-11-05 from the Flutter mobile app final review._ A slow Cloudflare call blocks login.
Fix: an `http.Client` with a timeout and the request context.
File: [backend/auth/utils/captcha.go](backend/auth/utils/captcha.go)

**S24. App Check replay guard is per instance and sweeps O(n) at its cap**
_Added 2026-11-05 from the Flutter mobile app final review._ With several auth instances a token can be replayed once per instance. At 100k live entries every call sweeps under the lock (only verified tokens enter).
Fix: a shared store (Redis) behind `ReplayGuard`.
File: [backend/auth/utils/replay_guard.go](backend/auth/utils/replay_guard.go)

**S25. App Check JWKS refresh failures are invisible**
_Added 2026-11-05 from the Flutter mobile app final review._ The SDK refreshes keys internally with no error hook, so a JWKS outage after boot only shows as rejected tokens (logged at Warn).
Fix: a periodic health log or a metric of verification failures by cause.
File: [backend/auth/utils/appcheck.go](backend/auth/utils/appcheck.go)

**S26. Request bodies are unbounded in user-service handlers**
_Added 2026-11-05 from the Flutter mobile app final review._ Only one user handler wraps its body in `http.MaxBytesReader`.
Fix: `http.MaxBytesReader` in every handler, or a shared middleware.
File: [backend/user/handlers/device/register_device_private.go](backend/user/handlers/device/register_device_private.go) and most handlers under `backend/user/handlers/`

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

**L16. Account deletion for App Store guideline 5.1.1(v)**
_Added 2026-11-05 from the Flutter mobile app final review._ The backend only disables accounts (status `disabled`; mobile ports web's disable-account section). Apple requires in-app account deletion for apps that offer signup. Blocks the first App Store submission.
Fix: confirm whether disable satisfies review; otherwise add a delete-account flow (backend first, then web and mobile).
Files: `backend/user/services/user.go`, `backend/auth/services/auth.go`

**L17. A livestream VOD can stay `processing` for good**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ Transcode ends the livestream with `vodStatus: processing` and later marks the VOD `ready`/`failed` by the VOD id the livestream service returns. If that reply is lost (the 10 s timeout fires after the VOD was created, a dropped connection, a decode error), transcode never learns the id and the VOD stays `processing` and hidden. On `main` the same timeout still left a playable VOD, since the playback URL went in that one call. The same id dependency forces the deploy order vod → livestream → transcode: against an old livestream service the VOD is created `ready` with no playback URL.
Fix: settle the VOD by livestream id (the `vods.livestream_id` column is indexed): a vod internal route such as `PATCH /v1/internal/livestreams/{id}/vod/status`, and transcode stops depending on the returned id. Keep the 2 s / 5 s retries for the moment the VOD is not created yet.
Files: [backend/transcode/rtmp/finish.go](backend/transcode/rtmp/finish.go), [backend/vod/handlers/vod/update_vod_status_internal.go](backend/vod/handlers/vod/update_vod_status_internal.go)

---

### 🟡 MEDIUM

**L4. Inconsistent field name for message type across protocols**
The REST API uses `type?: DmMessageType`; the WebSocket type uses `messageType: DmMessageType`. Developers must track which name applies to which transport.
REST: [web/lib/api/dm.ts:110](web/lib/api/dm.ts#L110)
WebSocket: [web/types/dm.ts:94](web/types/dm.ts#L94)

**L5. `GetMessages` silently swallows errors**
The function catches all errors and returns `{ messages: [] }` rather than propagating the error. Callers cannot distinguish a successful empty result from a network/server failure. All other API functions in the same file return `ApiResponse<T>`.
File: [web/lib/api/chat.ts:5-16](web/lib/api/chat.ts#L5-L16)

**L9. DM push notifications ignore conversation mute**
`DmMessageService.Send` pushes every participant except the sender, including participants who muted the conversation (`isMuted` on the participant record). Muting today only affects the in-app UI, so a muted conversation still buzzes the phone.
Fix: skip participants with `IsMuted == true` when building the push recipient list in `pushDm`, plus a unit test.
Files: [backend/chat/services/dm_message.go:155](backend/chat/services/dm_message.go#L155), [backend/chat/domains/conversation.go:46](backend/chat/domains/conversation.go#L46)

**L10. A social link cannot be removed**
_Added 2026-11-05 from the Flutter mobile app final review._ The social-link validator rejects `""`, so the delete branch in `UpdateUser` is unreachable from web and mobile.
Fix: treat `""` as delete before validation.
Files: [backend/user/dto/social_media_links.go](backend/user/dto/social_media_links.go), [backend/user/services/user.go](backend/user/services/user.go)

**L12. Push queue drain at shutdown is unbounded**
_Added 2026-11-05 from the Flutter mobile app final review._ `Close` waits for every queued push; that can take queue × timeout / workers (about 5 min) and outlive the container stop grace.
Fix: drain with a deadline, then drop what is left.
File: [backend/user/push/async.go:107](backend/user/push/async.go#L107)

**L15. Pushes are not localized; `device_tokens.locale` is unused**
_Added 2026-11-05 from the Flutter mobile app final review._ Mobile sends the app language with each device registration and when it changes (`PushService.resendLocale`), but pushes are built from the stored notification text.
Fix: localize by device locale, or stop collecting the locale.
File: [backend/user/push/sender.go:47](backend/user/push/sender.go#L47) (`FromNotification`)

**L18. VOD master `CODECS` goes stale after a mid-stream encoder change**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ The recorder takes each rendition's `avc1` string from the first SPS only. When a streamer changes profile/level mid-stream (supported, ruling R10), the master keeps advertising the old one, and players that trust `CODECS` (Safari, MSE `isTypeSupported`) may pick a variant they then fail to decode. The hand-written `avcCodecString` / `Recorder.codecs` also duplicate gohlslib's `codecparams.Marshal`, already a dependency.
Fix: leave `CODECS` out for a rendition whose SPS changed partway (as before #6b), and build the strings with `codecparams.Marshal`.
File: [backend/transcode/live/recorder.go](backend/transcode/live/recorder.go) (`readSPS`, `avcCodecString`, `codecs`)

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

**L13. Image-only DM push detection depends on web's placeholder text**
_Added 2026-11-05 from the Flutter mobile app final review._ The chat service treats a message as image-only when its text matches `^Sent (an image|\d+ images)$`, the text web puts there.
Fix: decide by message type `image` with empty user text, sent explicitly by clients.
File: [backend/chat/services/dm_push.go:27](backend/chat/services/dm_push.go#L27)

**L14. Group DM push shows only the sender's name**
_Added 2026-11-05 from the Flutter mobile app final review._ A push from a group conversation does not say which group.
Fix: title "<group> · <sender>" for group conversations (product decision).
File: [backend/chat/gateway/userservice/push.go](backend/chat/gateway/userservice/push.go)

**L19. The final VOD status update retries permanent errors**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ `UpdateVODStatus` turns every non-2xx into an error, so `finishStream` retries a 404 (the owner deleted the processing VOD) or a 400 twice, waits 7 s and logs "gave up marking VOD …, it stays processing", which is wrong for a deleted VOD.
Fix: retry only transport errors and 5xx; log a 4xx once with its status.
Files: [backend/transcode/rtmp/finish.go](backend/transcode/rtmp/finish.go), [backend/transcode/gateway/livestream/http/http.go](backend/transcode/gateway/livestream/http/http.go)

**L20. The `with-user` media card still links an unready VOD**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ The owner's card variant hides the duration and the links for a VOD that is not `ready`, but the `with-user` variant still wraps the title in a link to the watch page. No list sends that variant an unready VOD today.
Fix: gate that title link on `isPlayable` like the other branches.
File: [web/components/livestream/media-card.tsx](web/components/livestream/media-card.tsx) (`renderContent`, with-user branch)

**L21. A slow session close marks a good recording `failed`**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ `finishStream` marks the VOD `failed` on any `Session.Close` error, including its 75 s timeout, while the shutdown keeps running and may still write the VOD master moments later. The recorder (60 s) and thumbnailer bounds fit inside 75 s, so this needs storage to be stuck. Accepted for now.
Fix (if it shows up): on a close timeout, wait for the session to finish in the background and settle the VOD then.
File: [backend/transcode/rtmp/finish.go](backend/transcode/rtmp/finish.go)

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
`/v1/internal/gifts/create` and `/v1/internal/inventory/add` on the user service have no auth; they are safe only because the gateway does not route them. An internal API key header would harden this.
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

### 🟢 Small fixes

**W7. `REFRESH_EXCLUDE_PATHS` lists `/auth/register`, which does not exist**
_Added 2026-11-05 from the Flutter mobile app final review._ The signup route is `/auth/signup`, so a 401 from it triggers a pointless refresh.
Fix: replace it with `/auth/signup` (what mobile excludes).
File: [web/utils/fetchClient.ts:120](web/utils/fetchClient.ts#L120)

---

## Mobile app

Moved to [MOBILE_ISSUES.md](MOBILE_ISSUES.md) (M1–M12).

---

## Design Consistency Issues

_Added 2026-09-08 from a full-repo review. Scope is **code vs. code design** — convention drift, duplicated or dead registries, half-finished migrations. Not behavioral bugs (see Logic/Consistency above) and not TODOs. Nothing here is a security issue._

### 🟠 HIGH — structural

**C8. `auth` was never migrated to the per-feature handler layout**
Every other service uses `handlers/<feature>/<one_op_per_file>_{public,private,internal}.go` with a shared `basehandler`. `auth` still has 4 flat files, no `basehandler` (its own package-level `writeResponse` instead), and no `Public/Private/Internal` handler suffixes. Its services layer is flat too.
Files: [backend/auth/handlers/](backend/auth/handlers/), [backend/auth/services/](backend/auth/services/)

**C9. `auth` has zero distributed tracing**
`tracer.MyTracer.Start` call sites per service: user 28, vod 17, finance 11, livestream 5, **auth 0**. Tempo tracing is stated architecture, so auth is a blind spot in every login/signup trace.

**C10. The NATS event bus is fully built and completely unwired**
[backend/shared/pkg/eventbus/](backend/shared/pkg/eventbus/) has producer, consumer, admin, event_builder and 7 domain event files (finance, livestream, notification, transcode, user, vod) — with **zero imports outside `shared/`**. A NATS container runs in [docker-compose.yaml](docker-compose.yaml) and no service is given `NATS_URL`. All inter-service calls go over HTTP gateways. [docs/NATS_SETUP.md](docs/NATS_SETUP.md) documents the unused path.

**C11. `user` is half-migrated**
Handlers are split per-operation, but [backend/user/services/](backend/user/services/) is still flat one-file-per-domain (`user.go` is 290 lines). It also has no `handlers/utils/query.go`: pagination is parsed inline and `limit` is unsupported, where peers use `utils.GetPageAndLimitQuery`.
File: [backend/user/handlers/user/get_recommended_channels_public.go](backend/user/handlers/user/get_recommended_channels_public.go)

---

### 🟡 MEDIUM

**C12. `basehandler` duplicated 4× instead of living in `shared/`**
finance/livestream/user/vod each carry a copy. user's has already drifted: `res interface{}` instead of `res any`, capitalized comments, trailing whitespace.
File: [backend/user/handlers/basehandler/basehandler.go](backend/user/handlers/basehandler/basehandler.go)

**C13. `gateway/utils.go` duplicated 5× byte-identical**
auth/finance/livestream/user/vod are identical; transcode's differs. Never extracted to `shared/`, unlike `response`.

**C14. `handlers/utils/cookie.go` duplicated 4× identical** modulo the import path.

**C15. `shared/response` delegation is only half-adopted**
auth/user/vod/finance re-export [backend/shared/response](backend/shared/response/response.go) via type aliases; **livestream and transcode still carry full local copies** of the same types.

**C16. `auth`'s response package is named `serviceresponse`, not `response`**
Directory is `response/` in all 7 services but auth declares `package serviceresponse`, which also forces a different import alias at all 27 of its call sites.

**C17. Gateway naming: `user` vs `userservice`**
finance uses `gateway/userservice/` + `UserServiceGateway`; auth/livestream/vod use `gateway/user/` + `UserGateway`.

**C18. Gateway DTO placement is inconsistent**
auth = `gateway/user/dto/` (3 files); transcode = `gateway/user/dto.go` (a file) *and* `gateway/livestream/dto/dto.go` (a package); finance = none. transcode's `gateway/user/` also declares no gateway interface, unlike every other gateway.

**C19. The frontend runs two data-fetching architectures at once**
Newer domains (wallet, shop, notifications, dm, vod-comments, channels) use react-query hooks + `unwrapResponse`/`ApiError` (throw). Older ones use `useEffect` + manual `.success` checks: [user profile page](<web/app/[lng]/(main)/users/[userId]/page.tsx>), [settings/vods/page.tsx](<web/app/[lng]/(main)/settings/vods/page.tsx>), [_components/header/search.tsx](<web/app/[lng]/(main)/_components/header/search.tsx>), [wrappers/UserInformationWrapper.tsx](web/components/wrappers/UserInformationWrapper.tsx), [livestream/media-card.tsx](web/components/livestream/media-card.tsx), + 4 more. Note all of [web/lib/api/](web/lib/api/) returns `ApiResponse<T>`, so the throw convention exists only in the hook layer.

**C20. Query-cache helpers live in two places**
[web/lib/query/dm-cache.ts](web/lib/query/dm-cache.ts) and `vod-comments-cache.ts` vs [web/lib/api/user-cache.ts](web/lib/api/user-cache.ts).

**C21. `_components/` colocation is applied inconsistently**
Used in messages/notifications/settings/wallet, but [the user profile route](<web/app/[lng]/(main)/users/[userId]/>) puts `chat.tsx`, `gift-modal.tsx`, `profile.tsx`, `profile-header.tsx` next to `page.tsx`; same for `settings/vods/vod.tsx` and `wallet/inventory/send-gift-dialog.tsx`.

**C22. Route paths mix singular and plural**
The user service serves both `/v1/users/...` and `/v1/user/...`; vod serves `/v1/vods` and `/v1/vod-comments`.
File: [backend/user/api/http.go](backend/user/api/http.go)

**C23. One-operation-per-file is violated in a few places**
[backend/vod/services/vod_comment/like.go](backend/vod/services/vod_comment/like.go) holds both `LikeComment` and `UnlikeComment`; [backend/vod/handlers/general/general.go](backend/vod/handlers/general/general.go) holds both routes where peers split them into `route_not_found.go` + `route_service_health.go`; `user/handlers/follow/` and `user/handlers/user/upload_single_file_to_minio.go` lack the visibility suffix.

**C24. `chat` diverges from the Go services' layering and from its own naming**
No repository layer — services hit mongoose models directly. Within `src/` it mixes three conventions at once: `services/chatCommandService.ts` (camelCase), `models/Conversation.ts` (PascalCase), `types/chat-event.ts` (kebab-case).

---

### 🟢 LOW

**C25. Frontend filename casing has four conventions**
kebab-case dominates; PascalCase islands in [web/components/forms/](web/components/forms/) and [web/components/wrappers/](web/components/wrappers/); camelCase in `utils/fetchClient.ts`, `uploadClient.ts`, `timeFormats.ts`, `lib/validations/changePassword.ts`, `signUp.ts`; snake_case in `components/custom_react_player/` and `_components/header/header_utils_for_non_logged.tsx`.

**C26. [web/hooks/user.ts](web/hooks/user.ts) is a zustand store** but is not named `use-*-store.ts` like `use-dm-store.ts` and `use-upload-store.ts`.

**C27. `web/constant/` is singular** where `contexts/`, `hooks/`, `types/`, `utils/` are plural.

**C28. The two TypeScript codebases have opposing prettier configs**
[web/.prettierrc.js](web/.prettierrc.js): `semi: true, singleQuote: false`. [backend/chat/.prettierrc](backend/chat/.prettierrc): `semi: false, singleQuote: true, printWidth: 120`.

**C29. Root [.prettierignore](.prettierignore) is stale** — ignores `web/` (which has its own config) and `mobile/`, which is not tracked in this repo at all.

**C30. `docs/openapi.yaml` is missing for vod and transcode** (auth/finance/livestream/user have one; chat keeps its spec at `src/docs/openapi.yaml`).

**C31. Test coverage is very uneven**
The whole Go backend has 4 `_test.go` files (all under finance + shared/natsbus). chat has jest and one test. **web has zero tests and no `typecheck` script**, despite CI.

**C32. First-migration filenames differ per service** — `0001_create_tables.sql` (auth) vs `0001_init_tables.sql` (finance, livestream) vs `0001_add_user_table.sql` (user) vs `0001_init_vod_tables.sql` (vod).

**C33. [README.md](README.md) still documents IPFS ports 4001/8888** in the PORTS section while stating IPFS is no longer supported.

**C34. `shared/middlewares/cors.go` is used only by transcode's webserver.** The 5 API servers apply only `LoggingMiddleware` + `RequestIDMiddleware`.

**C35. Response-package import aliases vary** — `serviceresponse "…/auth/response"`, bare `"…/vod/response"`, and redundant `response "…/vod/response"` all appear (50 sites use the redundant alias).

**C36. Web mock mode gained two fake VODs**
_Added 2026-11-09 from the second review of `feat/ll-hls-packager` (details in [the problems file](docs/superpowers/reviews/2026-11-09-ll-hls-pipeline-problems.md))._ To check the owner's "Processing" / "Processing failed" labels, two records (`vod-processing`, `vod-failed`) were added to the existing MSW mock data, and the mock `GET /vods/{id}` mirrors `VisibleTo`. The user-global rules forbid mock / fake data.
Decide: keep them (mock mode is opt-in local tooling that already existed) or remove them (the labels then have nothing to show in mock mode).
File: [web/mocks/db.ts](web/mocks/db.ts)

---

### Suggested order for the consistency work

1. **C10** — decide the event bus's fate: wire it or delete it. It is the largest piece of unused machinery in the repo and it shapes every later inter-service decision.
2. **C8 + C9** — bring `auth` up to the layout every other service uses, and instrument it. Auth is currently invisible in traces.
3. **C15 + C12 + C13 + C14** — finish the `shared/` extraction that `response` started; each duplicate is an independent drift source (C12 has already drifted).
4. **C11** — finish the `user` migration, including the pagination helper.
5. **C19 + C20** — pick one frontend data-fetching architecture and converge on it.
6. **C31** — add a `typecheck` script to web, then tests where they pay off.
7. Everything else is mechanical and safe to batch: C16–C18, C21–C30, C32–C35.

---

## Recommended Fix Order

1. **S8** — Rotate all leaked secrets immediately; remove `.env` from git history
2. **S2 + S3** — Enforce token revocation on logout and in the refresh flow
3. **S4** — Add auth middleware and MIME validation to the upload endpoint
4. **S5 + S16** — Fix CORS; add `SameSite=Strict` to cookies
5. **S7** — Set OTP rate limit to 1/min
6. **S6** — Enable TLS between the gateway and services
7. **F1 + F2** — Compensation on `context.WithoutCancel`; HTTP client timeout in finance→user gateway (before merging `feat/finance-service`)
8. **F3 + F4** — Check Stripe `payment_status` before crediting; guarded payment status transitions
9. **L1** — Send error events to WebSocket clients on validation failure
10. **L2 + L3** — Align backend type definitions with actual runtime payloads
