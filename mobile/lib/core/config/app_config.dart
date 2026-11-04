/// Build flavor of the app. Matches the Android productFlavors and iOS schemes.
enum Flavor { dev, staging, prod }

/// Thrown when required build-time configuration is missing or invalid.
class ConfigException implements Exception {
  const ConfigException(this.message);

  final String message;

  @override
  String toString() => 'ConfigException: $message';
}

/// Backend endpoints and flavor, built from env keys mirroring
/// `web/example.env` (`NEXT_PUBLIC_BACKEND_*`).
///
/// Required keys: BACKEND_PROTOCOL, BACKEND_WS_PROTOCOL, BACKEND_HOST,
/// BACKEND_PORT, FLAVOR.
class AppConfig {
  const AppConfig._({
    required this.apiBaseUrl,
    required this.wsBaseUrl,
    required this.flavor,
  });

  factory AppConfig.fromEnv(Map<String, String> env) {
    final protocol = _require(env, 'BACKEND_PROTOCOL');
    final wsProtocol = _require(env, 'BACKEND_WS_PROTOCOL');
    final host = _require(env, 'BACKEND_HOST');
    final portText = _require(env, 'BACKEND_PORT');
    final flavorText = _require(env, 'FLAVOR');

    final port = int.tryParse(portText);
    if (port == null || port < 1 || port > 65535) {
      throw ConfigException('BACKEND_PORT is not a valid port: "$portText"');
    }

    final flavor = Flavor.values.asNameMap()[flavorText];
    if (flavor == null) {
      throw ConfigException(
        'FLAVOR must be one of ${Flavor.values.map((f) => f.name).join(', ')}'
        ', got "$flavorText"',
      );
    }

    return AppConfig._(
      apiBaseUrl: '$protocol://$host:$port',
      wsBaseUrl: '$wsProtocol://$host:$port/realtime',
      flavor: flavor,
    );
  }

  /// Gateway origin without a path; Envoy rewrites `/` to `/v1/`.
  final String apiBaseUrl;
  final String wsBaseUrl;
  final Flavor flavor;

  static String _require(Map<String, String> env, String key) {
    final value = env[key]?.trim();
    if (value == null || value.isEmpty) {
      throw ConfigException('Missing required config value $key');
    }
    return value;
  }
}
