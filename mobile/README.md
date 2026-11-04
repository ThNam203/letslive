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
dart format --set-exit-if-changed lib test
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
