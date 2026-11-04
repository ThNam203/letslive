/// Compile-time environment, supplied via `--dart-define-from-file=env/<flavor>.json`.
///
/// `String.fromEnvironment` must be called with constant keys, so each key is
/// listed explicitly. FLAVOR falls back to the value Flutter injects when
/// building with `--flavor`.
const Map<String, String> compileTimeEnv = {
  'BACKEND_PROTOCOL': String.fromEnvironment('BACKEND_PROTOCOL'),
  'BACKEND_WS_PROTOCOL': String.fromEnvironment('BACKEND_WS_PROTOCOL'),
  'BACKEND_HOST': String.fromEnvironment('BACKEND_HOST'),
  'BACKEND_PORT': String.fromEnvironment('BACKEND_PORT'),
  'FLAVOR': String.fromEnvironment(
    'FLAVOR',
    defaultValue: String.fromEnvironment('FLUTTER_APP_FLAVOR'),
  ),
};
