# LetsLive mobile

Flutter client for LetsLive (iOS 15+, Android 8+/API 26+). Flavors: `dev`, `staging`, `prod`.
Bundle ID / application ID: `vn.letslive.mobile` (`.dev` and `.staging` suffixes).

## Configuration

Backend endpoints come from compile-time defines, never from code. Copy the
example in `.env.example` to `env/dev.json` (git-ignored, one file per flavor):

```json
{
  "BACKEND_PROTOCOL": "http",
  "BACKEND_WS_PROTOCOL": "ws",
  "BACKEND_HOST": "localhost",
  "BACKEND_PORT": "8000",
  "FLAVOR": "dev",
  "TURNSTILE_DEV_TOKEN": "dev-token"
}
```

The Android emulator reaches the host machine at `10.0.2.2`, not `localhost`.
If a value is missing the app shows a configuration error screen naming it.

### Auth: bot proof, Firebase and Google sign-in

`POST /auth/login` and `POST /auth/verify-email` need a bot proof. With
Firebase configured (`FIREBASE_*` keys) the app sends a fresh single-use
Firebase App Check token per request (`X-Firebase-AppCheck`); dev without
Firebase sends `TURNSTILE_DEV_TOKEN` instead. Staging and prod require
Firebase and `GOOGLE_SERVER_CLIENT_ID`. All keys are listed in `.env.example`.

Google sign-in uses `google_sign_in` 7 and posts the id token to
`POST /auth/google/mobile`. Setup that needs the Google Cloud / Firebase
credentials (not in the repo):

- `GOOGLE_SERVER_CLIENT_ID` = the web OAuth client id (the backend's
  `GOOGLE_OAUTH_CLIENT_ID`).
- Android: register each flavor's package name (`vn.letslive.mobile`,
  `.dev`, `.staging`) with its signing SHA-1 as an Android OAuth client.
- iOS: create an iOS OAuth client per bundle id, set `GOOGLE_IOS_CLIENT_ID`,
  and add its reversed client id (`com.googleusercontent.apps.<id>`) to
  `CFBundleURLSchemes` in `ios/Runner/Info.plist`; without it the Google
  sheet cannot return to the app.
- App Check: register the Android app (Play Integrity) and iOS app (App
  Attest) in Firebase; in dev, register the debug token the app logs.

## Run, analyze, test

```sh
flutter pub get
flutter run --flavor dev --dart-define-from-file=env/dev.json
flutter analyze
flutter test
dart format --set-exit-if-changed lib test tool contract_test integration_test
```

## App shell and tabs

Signed-in screens live in `AppShell` (`lib/core/router/app_shell.dart`): the
current tab above a bottom `NavigationBar`. The tabs come from one registry,
`appTabs` in `lib/core/router/app_tabs.dart`, in display order (Home, Search,
Notifications, Messages, Profile). **To add a tab**, insert one `AppTab`
(`path`, `icon`, `label` from l10n, `builder`) at its place in that list;
the router branches and the bar are built from it, nothing else to wire. There
are no placeholder tabs: a tab appears when its screen exists, and the bar is
hidden while fewer than two tabs are registered. Each tab keeps its own
navigation stack. The shell also pauses the realtime socket in the background
and resumes it on return.

## Notifications and push

The Notifications tab (`lib/features/notifications/`) lists the user's
notifications and counts unread ones on its tab icon; new ones arrive live
over the realtime socket (`notification.created`). Push notifications
(`lib/core/push/`) reach the user while the app is closed or in the
background:

- **Registration.** After sign-in `PushService.registerForUser` asks for the
  notification permission and posts the FCM token to `POST /user/me/devices`
  (`{token, platform, locale}`); a rotated token is posted again. On logout
  and on session expiry the token is removed (`DELETE /user/me/devices/<token>`)
  and deleted locally, so a logged-out device gets no more pushes.
- **Taps.** A tap opens the screen the push names (`resolveDeepLink` in
  `lib/core/push/push_tap.dart`): a DM opens its conversation, a comment its
  video, anything else the notifications tab. Only the app's own routes are
  ever opened.
- **Foreground.** The server skips the push for a user with a live realtime
  connection, and the app shows nothing for a push that arrives while open
  (`setForegroundNotificationPresentationOptions` with everything off), so
  nothing shows twice: the in-app badge and list update instead.
- **Without Firebase** (dev with no `FIREBASE_*` keys) `PushService` does
  nothing and the rest of the app works as usual.

### Manual verification (needs a physical device)

Not done by the automated tests: it needs credentials the repository does not
have. Do this once per platform before a release.

What you need:

- A Firebase project with the Android and iOS apps registered, and its
  `FIREBASE_*` keys in `env/<flavor>.json` (see `.env.example`).
- iOS: an APNs authentication key (`.p8`) uploaded in Firebase console
  (Project settings, Cloud Messaging), a provisioning profile with the Push
  Notifications capability for the bundle id (the entitlement is in
  `ios/Runner/Runner.entitlements`), and a physical iPhone (the simulator has
  no APNs token).
- Android: a device or emulator with Google Play services, Android 13+ to
  see the permission prompt.
- The backend running with `FIREBASE_SERVICE_ACCOUNT_JSON` set (otherwise it
  logs pushes and sends none), reachable from the device.

Steps:

1. Run the app (`flutter run --flavor dev --dart-define-from-file=env/dev.json`
   on the device), sign in as user A and allow notifications.
2. Check the server has the device (this is also the first sign-in on a
   fresh install: the APNs token on iOS can arrive late, and the row must
   still appear without restarting the app):
   `SELECT platform, locale FROM device_tokens WHERE user_id = '<A id>'` in
   the user database shows one row.
3. Background the app (home button) or swipe it away. With the web app or
   another account B, follow user A, or send A a direct message, or comment on
   A's video. Expected: a system notification with the title and text of the
   in-app notification arrives within a few seconds (Android channel
   "Notifications" / "Thông báo").
4. Tap it. Expected: the app opens on the matching screen (DM: the
   conversation; comment: the video; follow or other: the Notifications tab),
   with back returning to the tabs. Repeat once with the app fully closed.
5. Keep the app open on screen and repeat step 3. Expected: no system
   notification; the Notifications (or Messages) tab badge goes up instead.
6. Delete the target (e.g. the video) before tapping a push for it. Expected:
   a "not found" notice with a back button.
7. Log out, then repeat step 3 for user A. Expected: no push. The
   `device_tokens` row of step 2 is gone.

Record the results in the pull request.

## Video player

Live (and later VOD) playback goes through one small interface,
`PlayerController` (`lib/features/watch/domain/player_controller.dart`);
`BetterPlayerControllerWrapper` wraps the package and widget tests use a fake.
The stream URL is built like web: `<api>/transcode/<livestreamId>/index.m3u8`.

Spike (2026-11-04; metadata from pub.dev / GitHub, packages read at source;
no physical device available):

| Criterion | `video_player` 2.14.1 alone | `better_player_plus` 2.0.1 |
|---|---|---|
| Background audio iOS | yes (`allowBackgroundPlayback` + `UIBackgroundModes`) | yes (sets `AVAudioSession` `.playback`; needs `UIBackgroundModes`) |
| Background audio Android | keeps playing, but no foreground service or media notification, so the OS may kill it | yes: `mediaPlayback` foreground service + notification |
| Picture-in-picture | **no API** (needs a second package) | yes, iOS (AVPictureInPictureController) and Android 8+ |
| HLS quality selection | yes since 2.14 (`selectVideoTrack`) | yes (parses the master playlist, `setTrack`) |
| Maintenance | Flutter team; last release 2026-10-02 | last release 2026-10-29, 5 releases in 3 months, 160/160 pub points, ~48 open issues |
| License | BSD-3-Clause | Apache-2.0 |
| Extra packages to add | a PiP package | none directly; new transitive ones: `wakelock_plus`, `package_info_plus`, `flutter_widget_from_html_core`, `html`, `csslib`, `dbus`, `petitparser`, `xml`, and SwiftPM `hyperoslo/Cache` |

**Pick: `better_player_plus`.** It is the only candidate that covers PiP
without adding another package. `media_kit` was not shortlisted: its last
release was 2025-12 (older than 6 months), and it needs `media_kit_video`
plus `media_kit_libs_*`.

Platform config for background audio and PiP:

- iOS `Info.plist` and `Info-Dev.plist`: `UIBackgroundModes` = `audio`. Both
  background audio and PiP need it. The package sets the `AVAudioSession`
  category itself.
- Android `MainActivity`: `android:supportsPictureInPicture="true"`. The
  package's own manifest adds the foreground service and its permissions.
- Android `activity-alias` `${applicationId}.PlaybackActivity`: the
  notification tap opens `<applicationId>.<activityName>`, which with the
  `.dev`/`.staging` suffixes would name a class that does not exist.
- Android `build.gradle.kts`: pins the package's Kotlin JVM target to 17. On
  AGP 9 the package expects built-in Kotlin, which this project keeps off,
  so its Kotlin would otherwise target JDK 21 against Java 17 and fail.

The package only holds a wakelock on its fullscreen route, so the app adds
`wakelock_plus` directly (approved 2026-11-04): `LivePlayerView` keeps the
screen on while the stream plays and releases it on pause, error, end and
when the view goes away (leaving the screen, logout).

Known gaps:

- Android picture-in-picture is button-only (the package does not hook
  `onUserLeaveHint`, so going home does not enter PiP). It goes through
  the package's fullscreen route: landscape, no controls.

Still needs a physical device: background audio with the screen locked;
PiP enter/exit (Android: button only) and returning to the app, including
orientation and controls after leaving PiP; lock screen / notification
controls; quality switches on a real network; the screen stays on while
watching (and sleeps again after pause); Android 13+ notification permission (without it the
notification is hidden, but playback goes on).

`integration_test/live_player_test.dart` drives the real player on a
simulator or emulator against any live HLS ladder, for example a local
ffmpeg one (1080/720/360, served with `python3 -m http.server 8899`):

```sh
ffmpeg -re -f lavfi -i testsrc2=size=1920x1080:rate=10 -f lavfi -i sine \
  -filter_complex "[0:v]split=3[a][b][c];[a]scale=1920:1080[v1];[b]scale=1280:720[v2];[c]scale=640:360[v3]" \
  -map "[v1]" -map "[v2]" -map "[v3]" -map 1:a -map 1:a -map 1:a \
  -c:v libx264 -preset ultrafast -g 20 -c:a aac -f hls -hls_time 2 -hls_list_size 10 \
  -hls_flags delete_segments -master_pl_name index.m3u8 \
  -var_stream_map "v:0,a:0 v:1,a:1 v:2,a:2" "%v/stream.m3u8"
flutter test integration_test/live_player_test.dart -d <device> --flavor dev \
  --dart-define-from-file=env/dev.json \
  --dart-define=LIVE_PLAYER_TEST_URL=http://<127.0.0.1 or 10.0.2.2>:8899/index.m3u8
```

Keep the encoder at real time (light preset, low frame rate): a ladder that
lags behind the wall clock makes ExoPlayer fail with
`BehindLiveWindowException`.

## Creator tools and video upload

`lib/features/creator/` holds web's creator settings, reached from
Profile tab > settings: stream key and stream information, chat commands,
the creator's own videos (edit, delete) and video upload.

- **Stream key** is a secret: masked until revealed, never logged, never part of
  an error message. Generating a new key asks for confirmation first because
  the old key stops working at once.
- **Uploads are foreground only.** `UploadManager` (`upload_manager.dart`) is an
  app-wide queue (not `autoDispose`, like web's upload store): leaving the
  upload screen does not cancel anything, logout cancels everything. It keeps
  the screen awake while a file is being sent (`wakelock_plus`) and streams the
  file from disk, so a 100 MB video is never held in memory.
- **Interrupted uploads are not resumed.** The server has no resumable upload.
  If the app is backgrounded or killed and the request fails, the item shows
  an error with **Retry**, which sends the whole file again. A token that
  expires before or during the request is refreshed once and the request
  replayed by the auth interceptor (the file is re-read from disk).
- After sending, the manager polls `GET /vods/<id>` every 5 s (the transcode
  worker's own poll interval) until the video is `ready` or `failed`, and
  stops when the item is dismissed. Web does not poll.
- Format (`.mp4 .mov .avi .mkv .webm`) and size (100 MB) are checked before
  sending; the limits are in `FieldLimits`. The picker is the system video
  picker (`image_picker`); files only reachable from the Files app are not
  offered (no `file_picker`).
- Direct-message pictures use the same general upload (`UploadFileRepository`,
  `POST /upload-file`) as VOD thumbnails.

## Contract tests (real backend)

Contract tests parse real backend responses with the app's models, so API
drift fails the run. They are opt-in: plain `flutter test` only runs `test/` and
needs no backend. They live in `contract_test/` (not `integration_test/`,
which Flutter only runs on a device) and run headless on the host VM.

### 1. Bring up the dev stack

Prerequisites: Docker, a root `.env` (ignored by version control; copy
`example.env`, generate `ACCESS_TOKEN_PRIVATE_KEY` as described in the root
README DEPLOYMENT section, use the Turnstile test secret
`1x0000000000000000000000000000000AA`; `GMAIL_APP_PASSWORD` can stay a dummy),
and any other LetsLive stack on this machine stopped (fixed ports and the
172.29.0.0/16 network).

The config server normally clones the private `letslive-configs` repo. Point
`LETSLIVE_CONFIGS_DIR` at a checkout of letslive-configs (add it to `.env` or
export it); it is mounted read-only. Compose fails with a clear message if it
is unset.

```sh
# from the repository root
docker compose -f docker-compose-dev.yaml -f mobile/tool/contract-stack.override.yaml \
  up -d configserver consul envoy auth user main_db nats minio livestream vod finance chat realtime chat_db
```

`mobile/tool/contract-stack.override.yaml` renames the containers to
`ll-mobile-*` (no clash with other checkouts), serves the config checkout and
aliases the database as `letslive-db` (the hostname the configs use). Wait until
`docker ps` shows the services up.

### 2. Run

Add `TURNSTILE_DEV_TOKEN` (any non-empty value; the dev stack's test secret
accepts it) to your `env/dev.json` (ignored by version control), next to the
`BACKEND_*` keys:

```sh
flutter test contract_test --dart-define-from-file=env/dev.json
```

Optional defines: `CONTRACT_DB_CONTAINER` (default `ll-mobile-db`),
`CONTRACT_DB_USER` (`postgres`), `CONTRACT_AUTH_DB` (`letslive_auth`).

`startContractContext()` (`contract_test/harness.dart`) creates two users with
unique emails through the real `POST /auth/verify-email` + `POST /auth/signup`
path and returns authenticated `ApiClient`s. The dev stack cannot send mail, but
the auth service stores the OTP before it tries, so the harness reads it with
`docker exec <db> psql` from `sign_up_otps`. It sends the Turnstile dev
token (`TURNSTILE_DEV_TOKEN`); the User-Agent no longer matters to the backend's
captcha check. Extra request headers (such as the App Check token) go through
`ApiClient(headerProvider: ...)` or the per-call `headers:` parameter.

### 3. Integration test on a simulator

`integration_test/auth_flow_test.dart` drives the real app (login with a
wrong and the right password, then logout) against the same stack. It needs
an existing user, because a device cannot read signup codes from the
database: create one through the web app or with `POST /auth/verify-email`,
the code from `sign_up_otps`, then `POST /auth/signup`.

```sh
flutter test integration_test/auth_flow_test.dart -d <simulator id> --flavor dev \
  --dart-define-from-file=env/dev.json \
  --dart-define=AUTH_FLOW_EMAIL=<email> --dart-define=AUTH_FLOW_PASSWORD=<password>
```

## Builds

```sh
flutter build apk --debug --flavor dev --dart-define-from-file=env/dev.json
flutter build ios --simulator --flavor dev --no-codesign --dart-define-from-file=env/dev.json
```

## Network security

Only the dev flavor allows non-HTTPS traffic (Android `usesCleartextTraffic`
in `android/app/src/dev/AndroidManifest.xml`; iOS `NSAllowsLocalNetworking` in
`ios/Runner/Info-Dev.plist`, used by the `*-dev` build configurations).
Staging and prod keep the strict platform defaults.

## Android release signing

Staging and prod release builds (`flutter build apk|appbundle --release --flavor prod`)
are signed with a keystore described in `android/key.properties` (git-ignored).
The build fails with a clear message if the file is missing; it never falls
back to the debug key. Required keys:

```properties
keyAlias=<alias>
keyPassword=<key password>
storeFile=<path to the .jks, relative to android/>
storePassword=<keystore password>
```

The dev flavor release build uses the debug key.
