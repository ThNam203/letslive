# LetsLive mobile app: handoff notes

Status as of 2026-11-05, for the owner to review later. The plan is
`2026-11-04-flutter-mobile-app.md` and the spec is
`../specs/2026-11-04-flutter-mobile-app-design.md`.

## State

- Branch `feat/mobile-app` has 58 commits off `main`, covering all 20 plan
  tasks plus a final whole-branch review and its fix wave.
- Verified on the final commit:
  - `flutter analyze` is clean and format reports 0 changed.
  - `flutter test` passes 718.
  - `parity_check --release` and `sync_locales --check` pass.
  - The contract suite passes 64 against the local `ll-mobile` stack.
  - `go vet` and `go test` pass in backend/auth, backend/user and backend/chat.
  - The `auth_flow` and `live_player` integration tests pass on an iPhone 17
    Pro simulator.
  - The dev APK and the iOS simulator build compile.
- Coverage when last measured: data 84%, state/providers 92%, lib 83%.
- Release steps, required secrets and the device checklist are in
  `mobile/README.md` under "Release checklist". Backend gaps are in
  `docs/ISSUES.md`. Web/mobile differences are in `docs/PARITY.md`.

## What needs the owner

### Credentials (none are in the repo)

- A Firebase project, its service account, and the App Check registrations
  (Play Integrity, App Attest).
- An APNs key.
- The App Attest capability on the bundle id and the provisioning profile.
- Google OAuth client IDs (web, iOS, Android for each flavor's package and
  SHA-1). The iOS reversed client id must also go into `CFBundleURLTypes`.
- The Android upload keystore.
- `WEB_BASE_URL` for staging and prod.
- An App Store Connect API key.
- The GitHub environments `staging` and `prod` with their secrets. Add
  reviewers or a branch rule on `prod`. Each prod run asks for two approvals,
  one per job.

### Physical-device checks (not done, not claimed)

Run these on a real iPhone (iOS 15+) and a real Android phone (8+); the README
checklist has the detail.

- Background audio with the screen locked.
- Picture-in-picture enter and exit (Android only enters it from the button).
- Landscape: the player goes full screen.
- Kill and relaunch restores the session.
- Chat and DMs recover after airplane mode.
- Push: foreground, background, tapping through, and the first sign-in on a
  fresh install.
- iOS App Check sign-in (release builds).
- The photo and video pickers, including what happens when photo access is
  denied.
- Uploads with the app sent to the background.
- Whether iOS hands back HEIC or JPEG from the camera roll (profile pictures
  accept only JPEG and WebP).
- The release workflow has never run. iOS signing (through
  `Signing.xcconfig`) and the AAB path are untested with real certificates.

### Product and App Store decisions

- **Sign-in on every screen.** Mobile requires sign-in everywhere, while web
  lets guests browse. Combined with the missing account deletion, App Store
  guideline 5.1.1(v) may object.
- **Account deletion.** The backend can only disable an account (mobile now
  ports web's "disable account"). Apple requires in-app deletion; see ISSUES
  L16.

### Local machine

- Your own docker stack was stopped with your approval. Restart it with
  `docker compose -p letslive up -d` from `/Users/ictsaigon.vn/mywork/letslive`.
- The test stack containers (`ll-mobile-*`, compose project
  `feat-mobile-app`) are still running. Stop them when you no longer need
  them.
- Add `TRANSCODE_DB_USER` and `TRANSCODE_DB_PASSWORD` to your root `.env`
  (`example.env` now lists them).
- Your uncommitted `web/AGENTS.md` should point to the root `AGENTS.md`, which
  holds the parity rule.

### Backend gaps found during the build (all in `docs/ISSUES.md`)

- S20: DM `imageUrls` are not validated by the server.
- S21: `PATCH /auth/password` does not revoke refresh tokens.
- S22: Envoy does not normalize paths.
- S23: the Turnstile HTTP call has no timeout.
- S24: the App Check replay guard is per instance.
- S25: JWKS refresh failures are invisible.
- S26: request bodies are unbounded.
- L9: DM push ignores conversation mute.
- L10: a saved social link cannot be removed.
- L12: the push drain at shutdown is unbounded.
- L13: image-only DM detection depends on web's placeholder text.
- L14: group DM push titles name only the sender.
- L15: pushes are not localized.
- L16: account deletion.
- W7: web `REFRESH_EXCLUDE_PATHS` lists a nonexistent `/auth/register`.
- Mobile M1–M4: coarse `parity_check` accounting, plural forms not
  reconciled across locales, the duplicated `Info-Dev.plist`, and feed cards
  refetching their owner on scroll-back.

## Decisions made during the build

Each line gives the decision, then its cost if it turns out to be wrong.

### Project and tooling

- The branch was cut from `origin/main` because no `release` branch exists.
  Cost: a rebase later.
- `.claude/worktrees/` and `.superpowers/` are ignored through
  `.git/info/exclude`. Cost: none.
- Bundle id `vn.letslive.mobile`, with `.dev` and `.staging` suffixes. Cost:
  a rename before store registration.
- Configuration comes from `--dart-define-from-file` (git-ignored
  `env/<flavor>.json`), not dotenv. Firebase is configured through
  dart-defines only, with no `google-services.json` or plist. Cost: native
  files might be needed later.
- Cleartext and ATS exceptions are allowed in the dev flavor only. Staging and
  prod release builds fail without `key.properties`. Cost: extra local setup.
- `firebase_core` 4.x and `firebase_app_check` 0.4.x are pinned because 5.x
  conflicts with `google_sign_in_ios` under SwiftPM, and `firebase_messaging`
  stays on the matching line. Cost: a later major upgrade, done together.
- The SwiftPM `Package.resolved` files are committed. Cost: none.
- Player: `better_player_plus` (the only candidate with picture-in-picture
  and no extra package), plus a direct `wakelock_plus`, which you approved.
  Cost: a larger dependency surface.
- Models are hand-written strict parsers, not freezed or json_serializable.
  Cost: none (simpler, no dependency).
- The release pipeline uses no fastlane and no third-party upload actions. It
  uploads to TestFlight with `xcrun altool`, and the Android AAB is an artifact
  that you upload to Play by hand. Cost: one manual Play upload per release.
- Parity rows can be marked `planned` during the build, and `--release` fails
  while any remain. The rule lives in the root `AGENTS.md` only, not in
  `web/AGENTS.md`. Cost: you add one pointer line yourself.
- Device, account and store steps are a README checklist and are not claimed
  as done. Cost: bugs that only show on a device appear later.

### Contract tests and local stack

- Contract tests run headless in `mobile/contract_test/`. The harness reads
  the signup OTP from the dev Postgres through `docker exec psql`, which is
  test-only (no production bypass), and seeds rows with `runSql` where the API
  cannot create them. Test data stays in the dev databases. Cost: the harness
  is coupled to the dev DB schema.
- Your `letslive-configs` checkout is mounted read-only. Cost: dev configs
  follow that checkout's HEAD.
- The dev flavor sends the Turnstile test token through
  `TURNSTILE_DEV_TOKEN`, and the App Check header is sent whenever Firebase is
  configured. Cost: none.
- The transcode worker runs as `ll-mobile-transcode` in the test project (the
  base compose name clashed with your own container). Cost: none.

### Backend

- App Check replay protection is an in-memory single-use store per auth
  instance, behind a `ReplayGuard` interface, and mobile sends limited-use
  tokens. Cost: tokens can be replayed across instances until a shared store
  exists (S24).
- The App Check verifier is lazy and never blocks boot on JWKS. Cost: none.
- A refresh JWT that fails to parse or has expired now returns 401 (it was
  500), so the client can detect an expired session. Cost: small backend diff.
- The Google id token is URL-escaped and the request has a 10 s timeout.
  Cost: none.
- DM pushes go through a user-service internal endpoint that the chat service
  calls, and no notification row is created. Cost: an extra internal call.
- Presence for push suppression comes from NATS, and a NATS outage counts as
  offline (send the push). Cost: possible duplicate pushes during an outage.

### Session and state

- `SessionLifecycle` logout hooks run before the cookie jar is cleared; on
  expiry they are best effort and the local FCM token is deleted too. Logout
  uses `DELETE /auth/logout`. Cost: none.
- "No username" means null or an empty string. A logout button sits on the
  account-setup screen. Cost: none.
- Realtime never connects anonymously while logged in. It retries with
  backoff, and only `sessionExpired` stops it. Cost: some retries while
  offline.
- **Sign-in is required for every screen.** See above.
- Feature data providers use `autoDispose`. The app refetches on resume only
  after 30 s or more in the background. Overflow is tested at 320 px with
  text scale 2.0. Cost: a little less caching.
- DMs refetch on every reconnect (the socket closes in the background), and
  the list merges page 0. Cost: one request per app switch.

### Features

- Tab registry order: Home, Search, Notifications, Messages, Profile, with no
  placeholder tabs. Web's left bar (followed and recommended channels) became
  home sections. Cost: none.
- Home has one live list (`/popular-livestreams`), and VODs paginate. Cost:
  none.
- Search has one repository method per type and no refetch on resume.
  Highlighting folds Vietnamese accents only. Cost: accents in other languages
  are not folded in the highlight.
- Plan Task 13 was split into 13a (player and channel) and 13b (chat). Live
  watching lives on `/users/:userId`, like web. Cost: none.
- Stream end is detected by polling every 30 s. The player recovers once from
  a RangeError or `BehindLiveWindow` before it shows retry. Cost: up to 30 s to
  notice a stream has ended.
- Chat: the length limit counts UTF-16 units like the server, with a
  surrogate-safe truncate. `:name` emote autocomplete is included, and
  keyboard navigation is not (desktop only). Cost: the 500-unit edge case
  differs from web by one unit.
- VOD reuses the live player wrapper, and share is "copy link" (no
  `share_plus`). `WEB_BASE_URL` is required in staging and prod. Cost: no
  native share sheet.
- **Comment likes are optimistic and roll back on failure; web waits for the
  server.** Cost: a like may flash and then disappear.
- **Profile and background pictures accept JPEG or WebP only.** Cost: PNG
  photos are refused.
- Clearing a saved social link is refused with "Link can't be empty" (L10).
  Cost: links cannot be removed.
- Channel tabs are Chat (the default), Home, Videos and About, and the bio
  sits in About. Cost: a small layout change if you want it elsewhere.
- **Editing and deleting your own DMs works through a long press; web has no
  UI for it.** Mute and group management have no UI, as on web. Cost: removing
  one sheet.
- **Wallet and shop are reached from the Profile tab. There is no deposit
  entry point and no "buy on the web" text** (out of scope, and App Store
  anti-steering rules). Cost: users with no balance get no hint.
- **Uploads run only while the app is in the foreground,** with the screen
  kept awake through a reference-counted holder. DM pictures use web's
  `/upload-file`. Cost: long uploads need the app open.
- Phone validation checks only the E.164 shape (web also requires `+`).
  Rotating a live channel to landscape loses the chat draft. Cost: minor.
