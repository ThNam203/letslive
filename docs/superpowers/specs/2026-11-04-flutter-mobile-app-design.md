# LetsLive Mobile App (Flutter) — Design

Date: 2026-11-04
Status: draft for approval

## 1. Intent and why

LetsLive today has one client, the Next.js web app. Some users never open the
web version: they only want a phone app, or they dislike the web UI. The mobile
app is a **second front door for the same user base**, not a new audience and
not a new product.

Consequence: a person who only ever uses mobile must be able to do everything
a web user can do (except the items under non-goals). "Mimic every feature"
is a hard requirement, not a nice-to-have.

## 2. Goals

1. Flutter app for iOS and Android with feature parity to the current web app
   (inventory in section 4), in the **first release**, well tested.
2. Same direction as web: the mobile app tracks the web app. From now on a
   web feature change is not finished until mobile has the equivalent
   (section 8). This includes in-flight web work described in `docs/`
   (stories, channel badges/points) once it ships on web. *(This reading of
   "new changes must be made too" is the author's; confirm.)*
3. Push notifications (FCM for Android, APNs for iOS, via FCM) with a
   device-token registry on the backend. This is new backend work.
4. Use the existing backend and gateway as-is wherever possible. No
   mobile-only services, except what push notifications require.

## 3. Non-goals and anti-goals

Non-goals (first release):
- **Stripe deposits and payments are out of scope** (user decision). Wallet
  *viewing* is in. See section 4 for what that means for shop and gifts.
- Broadcasting from the phone camera. The web app has none (streamers use
  OBS over RTMP); mobile mirrors that. *(Author's reading; confirm.)*
- Tablet-specific layouts, Flutter web/desktop targets.
- Native in-app purchase.

Anti-goals (would be a failure even if it technically works):
- A mobile app that works only on a happy-path demo. Parity features that
  silently fail on real devices (background playback, reconnect, token
  expiry) count as failures.
- Mobile drifting from web after launch.
- Mock or fallback data of any kind. Errors show an error state with retry.

## 4. Feature inventory (parity target)

Source: web routes under `web/app`, web API clients under `web/lib/api`,
and the gateway routes in `configs/envoy/envoy.yaml`.

| Area | Web today |
|---|---|
| Auth | Email signup with email verification, login, Google OAuth, logout, update password, silent token refresh. Cloudflare Turnstile on login/signup. |
| Browse | Home feed (live, popular), VOD feeds, channel list in left bar |
| Search | Channels, live streams, VODs; accent-insensitive; All/Channels/Live/Videos tabs; infinite scroll; query in URL (see `docs/SEARCH.md`) |
| Watch | HLS live playback (adaptive 360/720/1080), VOD playback, timestamped text |
| VOD comments | List, create, edit, delete, like (liked-ids), timestamp links |
| Profile | Public user page, follow/unfollow, follower count, own-profile edit (profile picture, background, display fields), stream key reset |
| Live chat | Realtime messages, emote picker, slash commands (built-in + user/channel custom; see `docs/CHAT_COMMANDS.md`) |
| DMs | Conversations, unread counts, typing, read receipts, edit/delete message, add/remove participants |
| Notifications | List, unread count, mark all read, realtime updates |
| Wallet | Overview, transactions, inventory (**deposit page excluded**) |
| Shop and gifts | Shop items, purchase, inventory, send gift (modal on a profile), gifts received page |
| Creator | Stream settings (stream key, stream info), chat-commands settings, own-VOD management, VOD upload with progress manager (`docs/VOD_UPLOAD_FLOW.md`) |
| Cross-cutting | EN and VI locales, light/dark theme, cookie policy page |

Shop and gifts spend wallet balance. With deposits out of scope, users can
spend only balance they already have (for example from web). **Author's call:**
keep shop, gifts, and inventory in the first release, with the balance read
from the existing wallet API, and show the same insufficient-balance errors as
web. Deposit entry points are omitted, not stubbed.

Planned on web but not shipped (stories, channel badges/points, realtime
phases): treated as future parity work, not first-release scope, unless they
land on web before mobile launches.

## 5. Constraints and known traps (found in the code)

1. **The OpenAPI specs cannot be trusted.** Several are stale: for example
   `/livestreams/popular` in the spec vs `/popular-livestreams` in the
   gateway, `/auth/refresh` in the spec vs `/auth/refresh-token` in the web
   client, and VOD, shop, gifts, notifications are missing. The source of truth
   is, in order: (a) the gateway routes in `configs/envoy/envoy.yaml`, (b) the
   handlers and DTOs in `backend/*`, (c) the web API clients and types. Specs
   may be used as a hint only. A generated client must come from a **verified
   or regenerated** spec; the plan must include an audit-and-fix step for the
   specs (or a hand-written typed client) before anything depends on them.
2. **Auth is cookie-based.** Web sends `credentials: include`, refreshes on 401
   via `POST /auth/refresh-token`, and Google login is a browser redirect.
   Mobile needs a persistent cookie jar (or a backend change to also accept a
   bearer token); the choice is the builder's, with the preference to avoid
   backend changes if cookie handling is reliable in the chosen HTTP client.
3. **Turnstile captcha** gates login and signup (`backend/auth/utils/captcha.go`).
   Mobile needs a way through: embedded Turnstile widget in a webview, or a
   backend change to accept a mobile attestation. Builder proposes; this
   touches auth, so it needs explicit sign-off in the plan.
4. **Realtime** is one WebSocket at `/realtime` through the gateway using
   subscribe/unsubscribe/ping frames and event frames (`backend/realtime/protocol`).
   Mobile must handle backgrounding, reconnect with resubscribe, and
   idle timeout (gateway idle timeout is 3600s).
5. **Types differ from DTOs in places** (see `docs/API_CONTRACT_AUDIT.md`).
   Parse defensively at the boundary and fail loudly in tests when a contract
   changes.
6. **No push infrastructure exists** in the backend today.

7. **History and existing mobile hooks (found after approval).** A previous
   Flutter app existed and was removed in commit `091134ee` (2026-08-13,
   "diverged too far from web to maintain"). The user chose a **fresh
   rebuild**; the old app stays in git history as reference only
   (`git show 091134ee^:mobile/...`), including its
   `MOBILE_VS_WEB_MISMATCHES.md`. The backend still has mobile pieces:
   `POST /v1/auth/google/mobile` (Google ID token) is usable as-is, but login
   skips the captcha for any request whose User-Agent contains "dart",
   "flutter", "okhttp" or "letslive-mobile" (`isMobileClient`,
   `backend/auth/handlers/auth.go:40`). That is spoofable and must be replaced
   with app attestation (Firebase App Check). Cookie auth stays (Dio cookie
   jar); no bearer-token path is needed.

## 6. Decided by the user

- Flutter, best practices, feature parity with web, everything in release 1.
- Well tested.
- Stripe deposits/payments excluded for now.
- Implement FCM/APNs push with device tokens.
- Mobile follows web changes going forward.

## 7. Author's calls (the user did not decide these)

- **Location:** `mobile/` next to `web/` in this repo.
- **Stack:** Flutter stable, Riverpod for state, go_router, Dio with a cookie
  jar, `freezed`/`json_serializable` models, `better_player`/`media_kit`-class
  HLS player chosen by the builder after testing on devices, `flutter_localizations`
  with the existing EN/VI strings ported from `web/lib/i18n/locales`.
  Adding dependencies still goes through the normal ask-first rule.
- **Architecture:** feature-first folders, layered (data / domain / presentation),
  repository per backend area, one WebSocket manager shared by chat, DMs and
  notifications.
- **Testing:** unit tests for repositories and parsers, widget tests per screen
  (loading, error with retry, empty, success), integration tests for login,
  watch live, live chat, DM, follow, VOD upload, push tap-through. Target 80%
  coverage on the data and state layers. Contract tests that fail when a
  backend response stops matching the parsed model.
- **Push:** backend gets a device-token table (user id, token, platform,
  created/last-seen), register/unregister endpoints behind auth, and a sender
  hooked into the existing notification events. Tokens are removed on logout
  and on FCM "unregistered" responses. Tapping a push deep-links to the right
  screen.
- **Parity process:** a `docs/PARITY.md` checklist (web route/feature -> mobile
  screen -> tests), a CLAUDE.md/AGENTS.md rule that web feature changes need a
  mobile counterpart in the same ticket, and a CI check that fails when
  gateway routes or API types change without a mobile update.
- **Release channels:** TestFlight and Play internal testing first; store
  listing is out of this document.
- **OS targets:** iOS 15+ and Android 8+ (API 26+). Background audio and
  picture-in-picture for live and VOD playback are required.
- **Push scope:** push every notification type the web notification list
  shows; no per-type opt-out unless web has one (mirror web). Suppress the push
  when the user has an active realtime connection. DM pushes show sender name
  and message text (a lock-screen privacy setting can be added later).
- **Backend changes:** go through the normal branch/PR flow like any backend
  change (new migration for the device-token table, register/unregister
  endpoints, sender, gateway routes). Mobile uses build flavors (dev,
  staging, prod) with base URLs set per flavor.
- **Parity CI check:** blocking when gateway routes or API types change and
  neither `docs/PARITY.md` nor mobile code changed in the same PR; escape
  hatch is a PARITY.md entry recording "mobile intentionally differs".
- **Delivery:** release 1 is one scope as the user decided, but the plan is
  sequenced as vertical slices (auth, browse/watch, chat/DM, notifications and
  push, profile/social, shop/gifts/wallet view, creator tools) with internal
  builds along the way.

## 8. Staying in step with web

A web change is "done" only when one of these is true: mobile has the
equivalent, or a recorded decision says mobile intentionally differs.
Mechanisms: the parity checklist, the repo rule, and the CI check in section 7.

## 9. Left to the builder

Screen-level layouts (follow platform conventions, borrow web information
architecture), player package, HTTP/cookie package details, exact navigation
structure, deep-link scheme, accessibility details, crash reporting choice
(web uses Sentry), ordering of the work inside the plan.

## 10. Open items needing the user

Answered by the user:
- Backend services, including auth, may be changed. The plan should pick the
  cleanest mobile path (for example a bearer-token login, a mobile captcha
  path, and a mobile Google OAuth callback) and treat it as ordinary backend
  work, with the auth changes reviewed carefully since they touch security.
- The user will supply the Apple Developer, Firebase and Google Cloud
  credentials as needed. Never hardcode them: they go in env/secret storage,
  per environment. The plan should list exactly which credentials are needed
  and at which step, so they can be supplied just in time.

Still to confirm (author's readings, section 2 and 3): stories and badges
tracked as future parity; no phone-camera broadcasting.
- Confirmation of the author's readings in sections 2 and 3 (stories and
  badges tracked as future parity; no phone-camera broadcasting).
