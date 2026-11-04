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
  "FLAVOR": "dev"
}
```

The Android emulator reaches the host machine at `10.0.2.2`, not `localhost`.
If a value is missing the app shows a configuration error screen.

## Run, analyze, test

```sh
flutter pub get
flutter run --flavor dev --dart-define-from-file=env/dev.json
flutter analyze
flutter test
dart format --set-exit-if-changed lib test tool contract_test
```

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
