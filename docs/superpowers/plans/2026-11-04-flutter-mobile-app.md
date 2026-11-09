# LetsLive Flutter Mobile App Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a fresh Flutter app (`mobile/`) with feature parity to the web app, backed by FCM/APNs push, with a process that keeps mobile in step with web.

**Architecture:** Feature-first, layered Flutter app (data / domain / presentation) talking to the existing Envoy gateway (port 8000) with cookie auth, one shared WebSocket manager for chat, DMs and notifications. Backend gains: app-attestation captcha path, device-token registry, FCM sender hooked into notification creation.

**Tech Stack:** Flutter stable, Dart, Riverpod, go_router, Dio + cookie jar, freezed/json_serializable, Firebase (core, messaging, App Check), `google_sign_in`; Go services (auth, user) with Postgres migrations.

**Spec:** `docs/superpowers/specs/2026-11-04-flutter-mobile-app-design.md` (read it first; section 5 lists the traps).

## Global Constraints

- iOS 15+ and Android 8+ (API 26+); background audio and picture-in-picture for live and VOD playback.
- Flutter stable, Riverpod, go_router, Dio with a cookie jar, `freezed`/`json_serializable`. Location: `mobile/` next to `web/`.
- The OpenAPI specs are stale: the source of truth is, in order, `configs/envoy/envoy.yaml`, the handlers/DTOs in `backend/*`, then the web clients in `web/lib/api` and `web/types`. Never copy an endpoint from a spec or from the removed old app without checking it.
- No mock/fake data and no fallback to mock. On error show an error state with retry.
- Stripe deposits and payments are out of scope; omit deposit entry points (do not stub them).
- Secrets and Firebase/Apple/Google credentials live in env or secret storage per environment, never in code. The user supplies credentials; each task names exactly what it needs.
- Adding a dependency beyond the set approved in Task 1 requires asking the user first.
- TypeScript/Go/Dart: no `any`/`dynamic` escape hatches, no `// TODO` instead of handling, typecheck/analyze must pass.
- Coverage target 80% on the data and state layers.
- Every commit references this plan: end the message with `Refs: docs/superpowers/plans/2026-11-04-flutter-mobile-app.md`. Never use `--no-verify`.
- Branch from `release`, one feature = one PR (user git workflow).
- Commands: mobile tests `cd mobile && flutter test <path>`; analyze `cd mobile && flutter analyze`; Go tests `cd backend/<svc> && go test ./...`.

## Review Focus

Inputs and failures the spec implies but no feature task would otherwise test. Each has a test in the owning task.

1. Access token expires while a request, upload, or WebSocket is live: expected silent refresh once, then retry; if refresh fails, go to login without a loop (Tasks 2, 3, 19).
2. App goes to background or loses network and returns: expected resubscribe and refetch of missed chat/DM/notification state, no duplicate messages (Tasks 3, 13, 16, 17).
3. User with `username == null` (Google signup): expected forced account-setup, no access to other screens (Task 10).
4. Vietnamese text without accents, emoji, and very long strings in search, chat, and comments: expected accent-insensitive search, no layout overflow (Tasks 12, 13, 14).
5. Content that vanished (deleted VOD, ended stream, removed comment, push tapped for deleted item) or push received after logout: expected a not-found state with a way back, and no push to a logged-out device (Tasks 9, 14, 17).

## File Structure (new `mobile/`)

```
mobile/pubspec.yaml, analysis_options.yaml, l10n.yaml
mobile/lib/main.dart, app.dart
mobile/lib/core/config/app_config.dart        flavor + base URLs
mobile/lib/core/network/{api_client,api_response,api_endpoints,auth_interceptor}.dart
mobile/lib/core/realtime/{realtime_manager,realtime_frame}.dart
mobile/lib/core/router/app_router.dart
mobile/lib/core/theme/                          light/dark
mobile/lib/core/push/push_service.dart
mobile/lib/features/<feature>/{data,domain,presentation}/
mobile/l10n/app_en.arb, app_vi.arb               generated from web locales
mobile/tool/sync_locales.dart, tool/parity_check.dart
mobile/test/..., mobile/integration_test/...
docs/PARITY.md
backend/user/migrations/0012_add_device_tokens.sql
backend/user/{handlers,repositories,services}/device/, backend/user/push/
```

Features: `auth, home, search, watch (live), vod, profile, messages, notifications, wallet, shop, creator`.

---

## Phase 0: Foundations

### Task 1: Scaffold, flavors, tooling

**Files:** Create `mobile/` (via `flutter create`), `mobile/lib/core/config/app_config.dart`, `mobile/test/core/config/app_config_test.dart`, `.github/workflows/mobile.yml`.

**Interfaces:**
- Produces: `enum Flavor { dev, staging, prod }`; `class AppConfig { final String apiBaseUrl; final String wsBaseUrl; final Flavor flavor; const AppConfig.fromEnv(Map<String,String> env); }`. `apiBaseUrl` = `<protocol>://<host>:<port>/v1` mirroring `web/example.env` (`NEXT_PUBLIC_BACKEND_*`).

- [ ] **Step 1:** Install the Flutter stable SDK (not present on this machine; ask the user to run the install if sudo/GUI is needed), then `flutter doctor` and record the output.
- [ ] **Step 2:** Present the dependency list to the user for approval: riverpod (+annotation/generator/lint), go_router, dio, dio_cookie_manager, cookie_jar, flutter_secure_storage, shared_preferences, freezed(+annotation), json_serializable, web_socket_channel, video_player + a background/PiP-capable HLS player (builder picks after Task 13 spike), cached_network_image, image_picker, file_picker, google_sign_in, firebase_core, firebase_messaging, firebase_app_check, url_launcher, intl, mocktail, integration_test. Do not proceed until approved.
- [ ] **Step 3:** Write failing test `AppConfig.fromEnv builds http and ws urls from env` (assert `apiBaseUrl == "http://localhost:8000/v1"`, `wsBaseUrl == "ws://localhost:8000/v1/realtime"` given the web example env) and `AppConfig.fromEnv throws when host is missing`. Run: `flutter test test/core/config` → FAIL.
- [ ] **Step 4:** Implement `AppConfig` and dev/staging/prod flavors (Android flavors + iOS schemes). Run tests → PASS; `flutter analyze` → no issues.
- [ ] **Step 5:** Add CI workflow `mobile.yml` running analyze + test on `mobile/**` changes. Commit: `chore(mobile): scaffold flutter app with flavors and CI`.

### Task 2: API client, envelope, auth refresh

**Files:** `mobile/lib/core/network/{api_client,api_response,auth_interceptor,api_endpoints}.dart`; tests `mobile/test/core/network/*_test.dart`.

**Interfaces:**
- Produces: `class ApiResponse<T> { bool success; int statusCode; int code; String key; String message; T? data; Meta? meta; List<Map<String,dynamic>>? errorDetails; String requestId; }` (mirrors `web/types/fetch-response.ts`); `class ApiClient { Future<ApiResponse<T>> get<T>(String path, {Map<String,dynamic>? query, required T Function(Object?) parse}); post/put/patch/delete likewise; Future<ApiResponse<T>> upload<T>(...,{void Function(int sent,int total)? onProgress, CancelToken? cancel}); }`; `ApiEndpoints` constants (only those verified against envoy + web clients).
- Refresh: on 401 for non-auth paths call `POST /auth/refresh-token` exactly once concurrently (single-flight), retry the original request once.

- [ ] **Step 1:** Failing tests with a fake Dio adapter: `parses success envelope with meta`, `maps error envelope to ApiResponse(success:false) not exception`, `401 triggers one refresh for N concurrent requests then retries each once`, `refresh failure emits SessionExpired and does not loop`, `network timeout yields a retryable error`, `cookie jar persists across client instances` (Review Focus 1). Run → FAIL.
- [ ] **Step 2:** Implement as above; cookie jar persisted with secure storage-backed `PersistCookieJar`. Run → PASS.
- [ ] **Step 3:** Commit: `feat(mobile): api client with envelope parsing and single-flight refresh`.

### Task 3: Realtime manager

**Files:** `mobile/lib/core/realtime/{realtime_manager,realtime_frame}.dart`; tests under `mobile/test/core/realtime/`.

**Interfaces:**
- Produces: `class RealtimeManager { Unsubscribe subscribe(String topic); Unsubscribe onEvent(String type, void Function(Object? data) handler); Unsubscribe onReconnect(void Function() handler); Stream<RealtimeStatus> get status; void pause(); void resume(); }`. Behavior copied from `web/contexts/realtime-context.tsx`: ref-counted topics, frames `{op: subscribe|unsubscribe|ping}` in and `{op: event|error|pong, topic, type, data, code}` out (`backend/realtime/protocol/protocol.go`), reconnect backoff 1s doubling to 30s, resubscribe all topics on reconnect, event types from `web/constant/realtime.ts`.

- [ ] **Step 1:** Failing tests with a fake channel: `second subscriber to same topic sends no extra frame`, `last unsubscribe sends unsubscribe`, `reconnect resubscribes every topic and fires onReconnect`, `backoff doubles up to 30s`, `pause closes socket on background and resume reconnects`, `unauthorized close triggers refresh once then reconnect` (Review Focus 1, 2). Run → FAIL.
- [ ] **Step 2:** Implement with `web_socket_channel`; WS connects through the cookie jar's cookies (`wsBaseUrl`). Run → PASS.
- [ ] **Step 3:** Commit: `feat(mobile): shared realtime manager`.

### Task 4: i18n, theme, app shell wiring

**Files:** `mobile/tool/sync_locales.dart`, `mobile/l10n/app_{en,vi}.arb`, `mobile/lib/core/theme/*`, `mobile/test/tool/sync_locales_test.dart`.

**Interfaces:** Produces `dart run tool/sync_locales.dart [--check]` that converts `web/lib/i18n/locales/{en-US,vi-VN}/*.json` to ARB with namespaced keys (`<namespace>_<key>`); `--check` exits non-zero when ARB and web JSON differ. Produces light/dark `ThemeData` with a persisted theme mode provider `themeModeProvider`, and `localeProvider` persisted via shared_preferences.

- [ ] **Step 1:** Failing tests: `converts nested keys with namespace prefix`, `preserves interpolation placeholders`, `--check fails when a web key is missing in arb`. Run → FAIL.
- [ ] **Step 2:** Implement; run it once to generate ARB; `flutter gen-l10n`. Run tests → PASS.
- [ ] **Step 3:** Commit: `feat(mobile): locale sync from web, theming`.

### Task 5: Parity tooling and rule

**Files:** `docs/PARITY.md`, `mobile/tool/parity_check.dart`, `mobile/test/tool/parity_check_test.dart`, `.github/workflows/parity.yml`, edit `AGENTS.md` (repo root; create if absent) and `web/AGENTS.md`.

**Interfaces:** `dart run tool/parity_check.dart` (a) lists gateway path prefixes from `configs/envoy/envoy.yaml` and web API client paths from `web/lib/api/*.ts`, (b) fails if any is absent from both `mobile/lib/core/network/api_endpoints.dart` and the "intentionally differs" list in `docs/PARITY.md`, (c) fails if a web route under `web/app/(main)` has no row in `docs/PARITY.md`. `PARITY.md` table columns: `web route/feature | mobile screen | tests | status (done/differs/planned)`.

- [ ] **Step 1:** Failing tests on fixture inputs: `flags an envoy prefix missing from mobile and PARITY`, `accepts a prefix listed as intentionally differs`, `flags a web page with no PARITY row`. Run → FAIL.
- [ ] **Step 2:** Implement; seed `PARITY.md` with every web route from the spec section 4 as `planned` (deposit marked `differs: out of scope`). Add the rule to the AGENTS files: a web feature change is done only when mobile has the equivalent or PARITY.md records the difference. Run → PASS; wire CI to run on PRs touching `web/**`, `backend/**/api`, `configs/envoy/**`, or `mobile/**` (blocking).
- [ ] **Step 3:** Commit: `ci: add web/mobile parity check`.

### Task 6: Contract test harness against the real backend

**Files:** `mobile/integration_test/contract/harness.dart`, `mobile/test/README` section in `mobile/README.md`.

**Interfaces:** `Future<ContractContext> startContractContext()` creating two test users through the real signup path on the dev stack (`docker-compose-dev.yaml`), returning authenticated `ApiClient`s. Contract tests parse real responses with the app's models and fail on any parsing error, so API/DTO drift breaks CI (spec section 5.5).

- [ ] **Step 1:** Write `contract smoke: GET /user/me returns parseable user` using the harness; run against the dev stack → FAIL until harness exists.
- [ ] **Step 2:** Implement harness (needs dev stack up: `docker compose -f docker-compose-dev.yaml up -d`; credentials from `.env`). Run → PASS. Document how to run in `mobile/README.md`.
- [ ] **Step 3:** Commit: `test(mobile): contract test harness`.

---

## Phase 1: Backend work (can run in parallel with Phase 0 after Task 1)

### Task 7: Replace User-Agent captcha bypass with app attestation

**Files:** Modify `backend/auth/handlers/auth.go:40-65` (remove `isMobileClient` use), `backend/auth/utils/captcha.go`; create `backend/auth/utils/appcheck.go`, tests `backend/auth/handlers/auth_captcha_test.go`, `backend/auth/utils/appcheck_test.go`. Same check for signup verification (`RequestEmailVerificationHandler`, `backend/auth/dto/signup_verification_request.go`).

**Interfaces:**
- Consumes: Firebase project credentials from env `FIREBASE_PROJECT_ID`, `FIREBASE_SERVICE_ACCOUNT_JSON` (user supplies).
- Produces: `func VerifyAppCheck(ctx context.Context, token string) error` (verifies a Firebase App Check token via the Admin SDK or JWKS); handlers accept the header `X-Firebase-AppCheck` as an alternative to `turnstileToken`.

- [ ] **Step 1:** Failing Go tests: `login with User-Agent dart and no captcha is rejected` (the current bypass must be gone), `login with valid app check token skips turnstile`, `login with invalid app check token is rejected`, `web login still requires turnstile`. Run: `cd backend/auth && go test ./handlers ./utils` → FAIL.
- [ ] **Step 2:** Implement `VerifyAppCheck` and the handler change (ask the user to approve adding the Firebase Admin Go dependency before `go get`). Run → PASS.
- [ ] **Step 3:** Commit: `fix(auth): replace spoofable user-agent captcha bypass with app check`.

### Task 8: Device-token registry

**Files:** Create `backend/user/migrations/0012_add_device_tokens.sql`, `backend/user/domains/device.go`, `backend/user/repositories/device/*.go`, `backend/user/services/device/*.go`, `backend/user/handlers/device/*.go`; modify `backend/user/api/http.go`, `configs/envoy/envoy.yaml` (route prefix `/user/me/devices`), `backend/user/docs/openapi.yaml`; tests beside each.

**Interfaces:**
- Produces: table `device_tokens(id uuid pk, user_id uuid fk, token text unique, platform text check in ('ios','android'), locale text, created_at, last_seen_at)`; `POST /v1/user/me/devices {token, platform, locale}` (upsert by token, reassigns to the current user), `DELETE /v1/user/me/devices/{token}`; repository `ListTokensByUserID(ctx, userID uuid.UUID) ([]DeviceToken, error)` and `DeleteByToken(ctx, token string) error`.

- [ ] **Step 1:** Failing tests: `registering same token for another user reassigns it`, `invalid platform returns validation error`, `delete removes only caller's token`, `list returns tokens for user`. Run → FAIL.
- [ ] **Step 2:** Implement following the existing notification handler/repository layering; add the envoy route with the same auth filter as `/user/me/inventory`. Run → PASS.
- [ ] **Step 3:** Commit: `feat(user): device token registry`.

### Task 9: FCM sender on notification creation

**Files:** Create `backend/user/push/{sender,fcm}.go`; modify `backend/user/services/notification/` (create path used by `CreateNotificationInternalHandler`); tests `backend/user/push/sender_test.go`.

**Interfaces:**
- Consumes: Task 8 `ListTokensByUserID`, `DeleteByToken`.
- Produces: `type PushSender interface { Send(ctx context.Context, userID uuid.UUID, n Notification) error }`; `type PresenceChecker interface { IsOnline(ctx context.Context, userID uuid.UUID) (bool, error) }` (implementation reads the realtime service presence; builder picks internal HTTP or the existing NATS bus, following `docs/NATS_SETUP.md`). Payload fields: `title`, `message`, `type`, `referenceId`, `actionUrl` (used for deep links, Task 17). DM pushes include sender name and message text (spec section 7). Language from the token's `locale`.

- [ ] **Step 1:** Failing tests with fake FCM and presence: `sends to every token of the user`, `skips send when user is online over realtime`, `deletes token when fcm reports unregistered`, `send failure does not fail notification creation`, `no tokens means no call` (Review Focus 5: tokens removed on logout come from Task 8 delete). Run → FAIL.
- [ ] **Step 2:** Implement and call from the notification service after the row is created (best effort, logged on failure, never blocks). Ask the user to approve the FCM dependency and supply `FIREBASE_SERVICE_ACCOUNT_JSON` per environment. Run → PASS.
- [ ] **Step 3:** Commit: `feat(user): send fcm push on notification creation`.

---

## Phase 2: Mobile features (vertical slices; each ends with widget tests for loading/error-with-retry/empty/success, a contract test, and a PARITY.md row set to done)

Each task mirrors the named web API client and page; take endpoints from `configs/envoy/envoy.yaml` and the web client, then confirm with a contract test.

### Task 10: Auth and session

**Files:** `mobile/lib/features/auth/{data,domain,presentation}/*` (login, signup + email verification, account-setup, session gate), `mobile/test/features/auth/`, `mobile/integration_test/auth_flow_test.dart`. Reference: `web/lib/api/auth.ts`, `web/components/forms/*`, `web/app/(auth)/*`.

**Interfaces:**
- Consumes: `ApiClient` (Task 2), Task 7 App Check header (attach `X-Firebase-AppCheck` to auth calls from `firebase_app_check`).
- Produces: `class AuthRepository { Future<ApiResponse<void>> login(String email,String password); signup/verify...; Future<ApiResponse<void>> loginWithGoogle(String idToken) /* POST /auth/google/mobile {idToken} */; Future<void> logout(); }`; `sessionProvider` (`AsyncNotifier<Session?>`) that loads `GET /user/me`; router redirect: logged-out → `/login`, `username == null` → `/account-setup`.

- [ ] **Step 1:** Failing tests: `login stores session and navigates home`, `wrong password shows server message with retry`, `google sign in posts id token`, `user with null username is redirected to account-setup from any route` (Review Focus 3), `logout clears cookie jar and unregisters device token` (Review Focus 5), `signup validates same field limits as web` (limits from `web/lib/validations/*`). Run → FAIL.
- [ ] **Step 2:** Implement; Google client IDs and bundle/package registration are supplied by the user (needs Google Cloud credentials). Run tests → PASS; run `auth_flow_test` on a simulator against the dev stack.
- [ ] **Step 3:** Commit: `feat(mobile): authentication and session`.

### Task 11: App shell and home feed

**Files:** `mobile/lib/features/home/*`, `mobile/lib/core/router/app_router.dart`. Reference: `web/app/(main)/page.tsx`, `web/components/livestream/*`, `web/components/leftbar/*`, `web/lib/api/livestream.ts`.

**Interfaces:** Produces `HomeRepository { Future<ApiResponse<List<Livestream>>> live(); popularLive(); popularVods(); }`, bottom-nav shell (Home, Search, Notifications, Messages, Profile) and `Livestream`, `Vod` models matching `web/types/livestream.ts`, `web/types/vod.ts`. Owner info is fetched separately as web does.

- [ ] **Step 1:** Failing tests: `home shows live and popular sections`, `empty feed shows empty state`, `error shows retry that refetches`, `pull to refresh reloads`. Run → FAIL.
- [ ] **Step 2:** Implement; Run → PASS; contract test for each list endpoint.
- [ ] **Step 3:** Commit: `feat(mobile): app shell and home feed`.

### Task 12: Search

**Files:** `mobile/lib/features/search/*`. Reference: `docs/SEARCH.md`, `web/lib/search.ts`, `web/lib/api/*` search calls, `web/components/search/*`.

**Interfaces:** Produces `SearchRepository { search({required String q, SearchType type, int page}) }`, tabs All / Channels / Live / Videos, 5 per type on All, infinite scroll on type tabs, highlight of matched text, debounced input (`web/hooks/use-debounced-value.ts` delay).

- [ ] **Step 1:** Failing tests: `All tab caps at 5 per type`, `type tab loads next page on scroll end`, `accent-less query "tro choi" renders server results with highlight` (Review Focus 4), `very long query does not overflow`. Run → FAIL; implement; PASS; contract test with accent-less query against the real search endpoint.
- [ ] **Step 2:** Commit: `feat(mobile): search`.

### Task 13: Live watch, live chat, emotes, slash commands

**Files:** `mobile/lib/features/watch/*`, `mobile/lib/core/chat_parser/{command,emote}.dart`. Reference: `web/components/custom_react_player/*`, `web/app/(main)/users/[userId]/page.tsx`, `docs/CHAT_COMMANDS.md`, `web/lib/api/{chat,chat-command}.ts`, `web/components/{emote-picker,chat-command-suggestions}.tsx`.

**Interfaces:**
- Consumes: `RealtimeManager` (topic `room:<streamerId>` via `realtimeRoomTopic`, events `chat.message`, `member.joined`, `member.left`).
- Produces: `LivePlayerController` (HLS adaptive 360/720/1080 with quality selection, background audio, PiP); `ChatController { Stream<List<ChatMessage>> messages; Future<void> send(String text); }`; `parseChatInput(String input, {required List<ChatCommand> commands}) -> ParsedInput` with built-ins identical to web (`/shrug`, `/roll`, ...).

- [ ] **Step 1:** Spike: evaluate two HLS players on a physical iOS and Android device for background audio + PiP + quality switching; record the pick and reason in `mobile/README.md`. (Ask before adding the chosen package if not in the Task 1 list.)
- [ ] **Step 2:** Failing tests: `parser expands built-in commands exactly as web`, `custom channel command resolved from commands list`, `incoming chat.message appends once even after reconnect` (Review Focus 2), `emoji and vietnamese text render without overflow` (Review Focus 4), `stream ended shows ended state with back action` (Review Focus 5), `send failure keeps text and offers retry`. Run → FAIL; implement; PASS.
- [ ] **Step 3:** Integration test: login → open live stream → send chat message → message visible to a second client (contract harness). Commit: `feat(mobile): live watch and chat`.

### Task 14: VOD watch and comments

**Files:** `mobile/lib/features/vod/*`. Reference: `web/app/(main)/users/[userId]/vods/[vodId]/page.tsx`, `web/components/vod-comments/*`, `web/components/vod/timestamped-text.tsx`, `web/lib/api/{vod,vod-comment}.ts`.

**Interfaces:** Produces `VodRepository`, `VodCommentRepository { list, create, edit, delete, like, unlike, likedIds }`, `TimestampedText` widget that seeks the player when a `mm:ss` link is tapped.

- [ ] **Step 1:** Failing tests: `tapping 01:23 seeks to 83s`, `like is optimistic and rolls back on failure`, `edit and delete only offered on own comments`, `deleted vod shows not-found with back` (Review Focus 5), `long comment with emoji wraps` (Review Focus 4). Run → FAIL; implement; PASS; contract tests for each comment endpoint.
- [ ] **Step 2:** Commit: `feat(mobile): vod playback and comments`.

### Task 15: Profile, follow, own profile

**Files:** `mobile/lib/features/profile/*`. Reference: `web/app/(main)/users/[userId]/page.tsx`, `web/app/(main)/settings/profile/page.tsx`, `web/components/user/*`, `web/lib/api/user.ts`, `web/lib/validations/*`.

**Interfaces:** Produces `UserRepository { get(id), me(), updateMe(fields), uploadProfilePicture(File), uploadBackground(File), follow(id), unfollow(id) }`; follow button is optimistic with rollback; field limits shared from one `FieldLimits` constants file matching `web/constant/*`.

- [ ] **Step 1:** Failing tests: `follow updates count immediately and rolls back on error`, `username below 6 chars is rejected locally with web's message`, `image upload sends multipart and shows progress`. Run → FAIL; implement; PASS; contract tests.
- [ ] **Step 2:** Commit: `feat(mobile): profiles and follow`.

### Task 16: Direct messages

**Files:** `mobile/lib/features/messages/*`. Reference: `web/app/(main)/messages/**`, `web/hooks/{use-dm-realtime,use-dm-store,use-conversation-chat}.ts`, `web/lib/api/dm.ts`, `web/types/dm.ts`.

**Interfaces:** Produces `MessageRepository` covering `/conversations` list/create/update/leave, participants add/remove, messages list (cursor)/send/edit/delete, `read`, `typing`, `unread-counts`; DM realtime via `RealtimeManager` (same topics/events as `use-dm-realtime.ts`); unread badge provider `unreadDmCountProvider`.

- [ ] **Step 1:** Failing tests: `messages page by cursor without duplicates`, `typing indicator expires after the same timeout as web`, `reconnect refetches the open conversation and merges without duplicates` (Review Focus 2), `opening a conversation marks it read and decrements badge`, `edit/delete only on own messages`. Run → FAIL; implement; PASS; two-client integration test via the harness.
- [ ] **Step 2:** Commit: `feat(mobile): direct messages`.

### Task 17: Notifications and push client

**Files:** `mobile/lib/features/notifications/*`, `mobile/lib/core/push/push_service.dart`, platform setup (`ios/Runner`, `android/app`). Reference: `web/lib/api/notification.ts`, `web/hooks/use-notification-realtime.ts`, `web/types/notification.ts`.

**Interfaces:**
- Consumes: Task 8 endpoints, Task 3 event `notification.created`.
- Produces: `NotificationRepository { list(page), unreadCount(), markRead(id), markAllRead(), delete(id) }`; `PushService { Future<void> init(); Future<void> registerForUser(); Future<void> unregister(); Stream<PushTap> taps; }` posting `{token, platform, locale}` to `/user/me/devices`; `resolveDeepLink(PushTap) -> String route` using `actionUrl`/`referenceId`/`type`, falling back to the notifications screen.

- [ ] **Step 1:** Failing tests: `realtime notification increments badge and prepends item`, `mark all read zeroes badge`, `token registered after login and refreshed on rotation`, `logout unregisters token before clearing session` (Review Focus 5), `tap on push for a deleted item lands on a not-found state with back` (Review Focus 5), `resolveDeepLink maps each notification type`. Run → FAIL; implement; PASS.
- [ ] **Step 2:** Device verification on a physical iOS and Android device (APNs key, FCM config files supplied by the user): app in background receives a push triggered by following the test user; app in foreground does not duplicate it. Record results in the PR.
- [ ] **Step 3:** Commit: `feat(mobile): notifications and push`.

### Task 18: Wallet (view), shop, inventory, gifts

**Files:** `mobile/lib/features/{wallet,shop}/*`. Reference: `web/app/(main)/wallet/{overview,transactions,inventory}`, `web/app/(main)/shop`, `web/app/(main)/users/[userId]/gifts`, `web/components/{shop,user/gift-modal.tsx}`, `web/lib/api/{wallet,shop,gift}.ts`. Excluded: `wallet/deposit` and `/deposits`, `/payments`.

**Interfaces:** Produces `WalletRepository { overview(), transactions(page), currencies() }`, `ShopRepository { items(), purchase(itemId) }`, `GiftRepository { send(...), received(userId), sent() }`, inventory list from `GET /user/me/inventory`.

- [ ] **Step 1:** Failing tests: `insufficient balance shows web's error key message`, `purchase refreshes balance and inventory`, `no deposit entry point exists in wallet overview`, `transactions paginate`. Run → FAIL; implement; PASS; contract tests.
- [ ] **Step 2:** Commit: `feat(mobile): wallet view, shop, gifts`.

### Task 19: Creator tools and VOD upload

**Files:** `mobile/lib/features/creator/*`. Reference: `web/app/(main)/settings/{stream,chat-commands,vods,upload}/page.tsx`, `web/components/upload-manager/*`, `web/hooks/use-upload-store.ts`, `docs/VOD_UPLOAD_FLOW.md`, `web/lib/api/vod.ts`.

**Interfaces:** Produces `StreamSettingsRepository` (stream key view/reset, stream info), `ChatCommandAdminRepository` (list mine, create, patch, delete), `MyVodsRepository` (list `/vods/author`, update, delete), `UploadManager { Stream<UploadState> enqueue(File f, VodMeta m); void cancel(String id); }` using `ApiClient.upload` with `disableTimeout` semantics (no 15s timeout), accepted extensions `.mp4 .mov .avi .mkv .webm`, polls VOD status until `ready`.

- [ ] **Step 1:** Failing tests: `rejects unsupported extension before upload`, `progress events reach 100 then status polling until ready`, `cancel aborts the request and removes the entry`, `token expiry mid-upload refreshes and the upload does not restart` or, if resume is impossible, `fails with a clear retry that re-sends` (Review Focus 1), `stream key reset requires confirmation`. Run → FAIL; implement; PASS.
- [ ] **Step 2:** Integration test: upload a small mp4 through the dev stack until `ready`. Commit: `feat(mobile): creator tools and vod upload`.

---

## Phase 3: Release readiness

### Task 20: Release pipeline and full verification

**Files:** `.github/workflows/mobile-release.yml`, `mobile/fastlane/*` (or documented manual steps), `docs/PARITY.md`, `mobile/README.md`.

**Interfaces:** Consumes everything above. Produces TestFlight and Play internal builds per flavor; every `PARITY.md` row `done` or `differs` with reason.

- [ ] **Step 1:** Run and record: `flutter analyze` (no issues), `flutter test --coverage` (>= 80% on `lib/**/data` and state providers; report actual numbers), `dart run tool/parity_check.dart` (pass), `dart run tool/sync_locales.dart --check` (pass), integration and contract suites against the dev stack, and the push device checks from Task 17.
- [ ] **Step 2:** Manual pass on a physical iOS 15+ and Android 8+ device: background playback, PiP, rotation, kill/relaunch session restore, airplane-mode recovery on chat/DM.
- [ ] **Step 3:** Set up the TestFlight and Play internal tracks (needs the user's Apple/Google accounts and signing credentials; supply just in time). Commit: `ci(mobile): release pipeline`.

---

## Self-Review

- **Spec coverage:** auth incl. Google (T10, T7), browse/search (T11, T12), live/VOD/comments (T13, T14), profile/follow (T15), chat/emotes/commands (T13), DMs (T16), notifications + push + device tokens (T8, T9, T17), wallet/shop/gifts without deposit (T18), creator/upload (T19), EN/VI and theme (T4), parity process (T5), contract tests (T6), flavors and OS targets (T1, T20), testing targets (all + T20).
- **Not in this plan by design:** stories, channel badges/points (join mobile when they ship on web, tracked via PARITY.md), phone-camera broadcasting, deposits/payments.
- **Open for the user during execution:** credentials per task (Firebase service account T7/T9; Apple/Google/Firebase config files T10/T17/T20); dependency approvals (T1, T7, T9).
- **Type consistency:** `ApiClient`/`ApiResponse` (T2), `RealtimeManager` (T3), `PushSender`/`PresenceChecker` (T9), `PushService` (T17), `AuthRepository` (T10) are used by name in later tasks.
