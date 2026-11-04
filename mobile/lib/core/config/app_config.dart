import 'package:flutter/foundation.dart';

/// Build flavor of the app. Matches the Android productFlavors and iOS schemes.
enum Flavor { dev, staging, prod }

/// Thrown when required build-time configuration is missing or invalid.
class ConfigException implements Exception {
  const ConfigException(this.message);

  final String message;

  @override
  String toString() => 'ConfigException: $message';
}

/// Firebase project values passed to `Firebase.initializeApp`. Read from env
/// instead of google-services.json / GoogleService-Info.plist, so every
/// environment is configured the same way.
class FirebaseConfig {
  const FirebaseConfig({
    required this.projectId,
    required this.apiKey,
    required this.messagingSenderId,
    required this.appId,
  });

  /// Firebase project id (FIREBASE_PROJECT_ID). Not secret; safe to show in
  /// error messages.
  final String projectId;

  /// Firebase Web/API key of the project (FIREBASE_API_KEY). It ships inside
  /// the app, but is still never logged or shown.
  final String apiKey;

  /// Firebase Cloud Messaging sender id (FIREBASE_MESSAGING_SENDER_ID), the
  /// project number; required by `FirebaseOptions`.
  final String messagingSenderId;

  /// The app id of the platform the app runs on (Android or iOS).
  final String appId;
}

/// OAuth client ids for native Google sign-in.
class GoogleSignInConfig {
  const GoogleSignInConfig({required this.serverClientId, this.iosClientId});

  /// The web OAuth client id. Google puts it in the id token's `aud`, which
  /// the backend checks against GOOGLE_OAUTH_CLIENT_ID.
  final String serverClientId;

  /// The iOS OAuth client id; set only when running on iOS.
  final String? iosClientId;
}

/// Backend endpoints, flavor and auth settings, built from env keys. Backend
/// keys mirror `web/example.env` (`NEXT_PUBLIC_BACKEND_*`); every key is
/// documented in `mobile/.env.example`.
///
/// Always required: BACKEND_PROTOCOL, BACKEND_WS_PROTOCOL, BACKEND_HOST,
/// BACKEND_PORT, FLAVOR. Staging and prod also require Firebase (bot proof
/// via App Check) and GOOGLE_SERVER_CLIENT_ID. Dev without Firebase requires
/// TURNSTILE_DEV_TOKEN instead.
class AppConfig {
  const AppConfig._({
    required this.apiBaseUrl,
    required this.wsBaseUrl,
    required this.flavor,
    this.firebase,
    this.googleSignIn,
    this.turnstileDevToken,
  });

  /// Throws [ConfigException] naming the first missing or invalid key.
  /// [platform] picks the platform-specific ids; it defaults to the running
  /// platform.
  factory AppConfig.fromEnv(
    Map<String, String> env, {
    TargetPlatform? platform,
  }) {
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

    final isIos = (platform ?? defaultTargetPlatform) == TargetPlatform.iOS;
    final firebase = _firebase(env, isIos: isIos);
    final googleSignIn = _googleSignIn(env, isIos: isIos);
    String? turnstileDevToken;
    if (flavor == Flavor.dev) {
      // Without App Check the dev stack's Turnstile test secret is the only
      // way past the bot check.
      if (firebase == null) {
        turnstileDevToken = _require(env, 'TURNSTILE_DEV_TOKEN');
      }
    } else {
      if (firebase == null) _missingForRelease('FIREBASE_PROJECT_ID', flavor);
      if (googleSignIn == null) {
        _missingForRelease('GOOGLE_SERVER_CLIENT_ID', flavor);
      }
    }

    return AppConfig._(
      apiBaseUrl: '$protocol://$host:$port',
      wsBaseUrl: '$wsProtocol://$host:$port/realtime',
      flavor: flavor,
      firebase: firebase,
      googleSignIn: googleSignIn,
      turnstileDevToken: turnstileDevToken,
    );
  }

  /// Gateway origin without a path; Envoy rewrites `/` to `/v1/`.
  final String apiBaseUrl;
  final String wsBaseUrl;
  final Flavor flavor;

  /// Null when FIREBASE_PROJECT_ID is unset (dev only).
  final FirebaseConfig? firebase;

  /// Null when GOOGLE_SERVER_CLIENT_ID is unset (dev only); the Google button
  /// is hidden then.
  final GoogleSignInConfig? googleSignIn;

  /// Cloudflare Turnstile test token; set only in dev without Firebase.
  final String? turnstileDevToken;

  static FirebaseConfig? _firebase(
    Map<String, String> env, {
    required bool isIos,
  }) {
    final projectId = _optional(env, 'FIREBASE_PROJECT_ID');
    if (projectId == null) return null;
    return FirebaseConfig(
      projectId: projectId,
      apiKey: _require(env, 'FIREBASE_API_KEY'),
      messagingSenderId: _require(env, 'FIREBASE_MESSAGING_SENDER_ID'),
      appId: _require(
        env,
        isIos ? 'FIREBASE_IOS_APP_ID' : 'FIREBASE_ANDROID_APP_ID',
      ),
    );
  }

  static GoogleSignInConfig? _googleSignIn(
    Map<String, String> env, {
    required bool isIos,
  }) {
    final serverClientId = _optional(env, 'GOOGLE_SERVER_CLIENT_ID');
    if (serverClientId == null) return null;
    return GoogleSignInConfig(
      serverClientId: serverClientId,
      // The iOS SDK needs its own client id; Android only uses the server id.
      iosClientId: isIos ? _require(env, 'GOOGLE_IOS_CLIENT_ID') : null,
    );
  }

  static Never _missingForRelease(String key, Flavor flavor) =>
      throw ConfigException(
        'Missing required config value $key (required for ${flavor.name})',
      );

  static String? _optional(Map<String, String> env, String key) {
    final value = env[key]?.trim();
    return value == null || value.isEmpty ? null : value;
  }

  static String _require(Map<String, String> env, String key) =>
      _optional(env, key) ??
      (throw ConfigException('Missing required config value $key'));
}
