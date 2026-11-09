/// App-wide singletons shared by every feature: config, the cookie jar that
/// holds the session, the HTTP client and the realtime socket.
library;

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/secure_cookie_storage.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';

/// The startup config. `main` overrides it with the parsed env; reading it
/// without an override is a wiring bug and fails loudly.
final appConfigProvider = Provider<AppConfig>(
  (ref) => throw StateError('appConfigProvider must be overridden.'),
);

/// The persistent, secure-storage backed cookie jar (ACCESS/REFRESH tokens).
final cookieJarProvider = Provider<CookieJar>((ref) => createSecureCookieJar());

/// The gateway client. Shares [cookieJarProvider] with the realtime socket.
final apiClientProvider = Provider<ApiClient>((ref) {
  final client = ApiClient(
    baseUrl: ref.watch(appConfigProvider).apiBaseUrl,
    cookieJar: ref.watch(cookieJarProvider),
  );
  ref.onDispose(client.dispose);
  return client;
});

/// The single realtime socket. `main` starts it once at boot; the session
/// restarts it on login and logout so it re-authenticates.
final realtimeManagerProvider = Provider<RealtimeManager>((ref) {
  final client = ref.watch(apiClientProvider);
  final manager = RealtimeManager(
    url: ref.watch(appConfigProvider).wsBaseUrl,
    cookieJar: ref.watch(cookieJarProvider),
    refreshSession: client.refreshSession,
    sessionExpired: client.sessionExpired,
  );
  ref.onDispose(manager.dispose);
  return manager;
});
