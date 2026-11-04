/// Gateway paths used by the app. Paths exclude `/v1`; Envoy rewrites `^/` to
/// `/v1/`. Only endpoints verified against `configs/envoy/envoy.yaml`, the
/// backend handlers and the web clients are listed. Keep one
/// `static const String name = '/path';` per line (parsed by parity_check).
abstract final class ApiEndpoints {
  static const String login = '/auth/login';
  static const String signup = '/auth/signup';
  static const String logout = '/auth/logout';
  static const String refreshToken = '/auth/refresh-token';
  static const String verifyEmail = '/auth/verify-email';
  static const String userMe = '/user/me';
}
