import 'dart:async';
import 'dart:developer' as developer;

import 'package:cookie_jar/cookie_jar.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';

/// Work to do before the session ends, while its cookies still authenticate
/// (e.g. unregistering the push token, Task 17).
typedef LogoutHook = Future<void> Function();

/// Owns what happens around a session change: logout hooks, the server
/// logout call, clearing the cookie jar and restarting the realtime socket.
/// [SessionNotifier] calls it and then publishes the new session state.
class SessionLifecycle {
  SessionLifecycle({
    required this._auth,
    required this._cookieJar,
    required this._restartRealtime,
  });

  /// How long one logout hook may take before logout moves on without it.
  static const Duration defaultHookTimeout = Duration(seconds: 5);

  final AuthRepository _auth;
  final CookieJar _cookieJar;
  final void Function() _restartRealtime;
  final List<LogoutHook> _logoutHooks = [];
  Future<void>? _ending;

  /// Runs [hook] at the start of every logout and session expiry.
  void addLogoutHook(LogoutHook hook) => _logoutHooks.add(hook);

  /// After a login: reconnect the socket so it authenticates as the user.
  void signedIn() => _restartRealtime();

  /// Ends the session: hooks (each awaited; a failure or timeout is logged
  /// and skipped), then `DELETE /auth/logout` when [notifyServer] (false on
  /// expiry, the server already rejected the session), then clears the jar
  /// and restarts the socket as a guest.
  ///
  /// Concurrent calls share one run, so a hook whose request triggers another
  /// session-expired event cannot start a second logout (no loop).
  Future<void> endSession({
    required bool notifyServer,
    Duration hookTimeout = defaultHookTimeout,
  }) => _ending ??=
      // Started in a microtask so `_ending` is set before the first hook
      // runs; a hook that re-enters then joins this run.
      Future<void>.microtask(() => _end(notifyServer, hookTimeout))
          .whenComplete(() => _ending = null);

  Future<void> _end(bool notifyServer, Duration hookTimeout) async {
    for (final hook in List.of(_logoutHooks)) {
      try {
        await hook().timeout(hookTimeout);
      } on Object catch (e, stack) {
        // A broken hook must never keep the user signed in.
        developer.log(
          'logout hook failed',
          name: 'SessionLifecycle',
          error: e,
          stackTrace: stack,
        );
      }
    }
    if (notifyServer) {
      // The result is ignored: the local logout happens even offline.
      await _auth.logout();
    }
    await _cookieJar.deleteAll();
    _restartRealtime();
  }
}
