/// Compile-time environment, supplied via `--dart-define-from-file=env/<flavor>.json`.
///
/// `String.fromEnvironment` must be called with constant keys, so each key is
/// listed explicitly. FLAVOR falls back to the value Flutter injects when
/// building with `--flavor`. Keys are documented in `mobile/.env.example`.
const Map<String, String> compileTimeEnv = {
  'BACKEND_PROTOCOL': String.fromEnvironment('BACKEND_PROTOCOL'),
  'BACKEND_WS_PROTOCOL': String.fromEnvironment('BACKEND_WS_PROTOCOL'),
  'BACKEND_HOST': String.fromEnvironment('BACKEND_HOST'),
  'BACKEND_PORT': String.fromEnvironment('BACKEND_PORT'),
  'FLAVOR': String.fromEnvironment(
    'FLAVOR',
    defaultValue: String.fromEnvironment('FLUTTER_APP_FLAVOR'),
  ),
  'TURNSTILE_DEV_TOKEN': String.fromEnvironment('TURNSTILE_DEV_TOKEN'),
  'FIREBASE_PROJECT_ID': String.fromEnvironment('FIREBASE_PROJECT_ID'),
  'FIREBASE_API_KEY': String.fromEnvironment('FIREBASE_API_KEY'),
  'FIREBASE_MESSAGING_SENDER_ID': String.fromEnvironment(
    'FIREBASE_MESSAGING_SENDER_ID',
  ),
  'FIREBASE_ANDROID_APP_ID': String.fromEnvironment('FIREBASE_ANDROID_APP_ID'),
  'FIREBASE_IOS_APP_ID': String.fromEnvironment('FIREBASE_IOS_APP_ID'),
  'GOOGLE_SERVER_CLIENT_ID': String.fromEnvironment('GOOGLE_SERVER_CLIENT_ID'),
  'GOOGLE_IOS_CLIENT_ID': String.fromEnvironment('GOOGLE_IOS_CLIENT_ID'),
  'WEB_BASE_URL': String.fromEnvironment('WEB_BASE_URL'),
};
