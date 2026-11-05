/// Push notifications on the device: registers the FCM token with the
/// backend for the signed-in user, and reports taps on pushes.
///
/// Flow:
/// - [init] runs once at boot (after Firebase): it starts listening for token
///   rotation and for taps.
/// - [registerForUser] runs whenever a user is signed in: it asks permission,
///   then posts the token. Token rotation posts the new token while a user
///   is registered.
/// - [unregister] is a logout hook (`SessionLifecycle.addLogoutHook`, so it
///   runs on session expiry too): it deletes the token on the server while
///   the session still authenticates, and always deletes the local FCM token
///   afterwards, so even a failed server call (expiry, offline) leaves a
///   logged-out device with no token to push to.
///
/// Without Firebase configuration (`FIREBASE_*` keys unset, dev only) there
/// is no messaging: every method does nothing, and [taps] never emits.
library;

import 'dart:async';
import 'dart:developer' as developer;

import 'package:letslive_mobile/core/push/device_repository.dart';
import 'package:letslive_mobile/core/push/push_messaging.dart';
import 'package:letslive_mobile/core/push/push_tap.dart';

/// See the library comment.
class PushService {
  PushService({
    required this._messaging,
    required this._devices,
    required this._platform,
    required this._locale,
  });

  final PushMessaging? _messaging;
  final DeviceRepository _devices;
  final String _platform;
  final String Function() _locale;

  // Single-subscription on purpose: a tap that launched the app arrives in
  // [init], before the router is up to listen, and is kept until it does.
  final StreamController<PushTap> _taps = StreamController<PushTap>();
  StreamSubscription<String>? _tokenRefreshes;
  StreamSubscription<Map<String, Object?>>? _tapData;
  bool _initialized = false;

  /// True between [registerForUser] and [unregister].
  bool _active = false;

  /// The token the server was last given successfully.
  String? _registeredToken;

  /// Taps on pushes, in order. Listen once.
  Stream<PushTap> get taps => _taps.stream;

  /// Starts listening for token rotation and taps; silences pushes while
  /// the app is open. Call once at boot, after `Firebase.initializeApp`.
  Future<void> init() async {
    final messaging = _messaging;
    if (messaging == null || _initialized) return;
    _initialized = true;
    _tokenRefreshes = messaging.tokenRefreshes.listen((token) {
      if (_active) unawaited(_register(token));
    });
    _tapData = messaging.tapData.listen(_onTapData);
    try {
      await messaging.silenceForeground();
      final launchData = await messaging.launchTapData();
      if (launchData != null) _onTapData(launchData);
    } on Exception catch (e, stack) {
      _log('init failed', e, stack);
    }
  }

  /// Registers this device for the signed-in user: asks for permission, then
  /// posts the token. A denied permission registers nothing (the pushes
  /// could not be shown). Failures are logged; the next sign-in or token
  /// rotation tries again.
  Future<void> registerForUser() async {
    final messaging = _messaging;
    if (messaging == null) return;
    _active = true;
    try {
      if (!await messaging.requestPermission()) return;
      final token = await messaging.token();
      // Logged out while the prompt was open: do not register.
      if (token == null || !_active) return;
      await _register(token);
    } on Exception catch (e, stack) {
      _log('registration failed', e, stack);
    }
  }

  /// Removes this device from the signed-in user, then drops the local FCM
  /// token even when the server call failed.
  Future<void> unregister() async {
    final messaging = _messaging;
    if (messaging == null) return;
    _active = false;
    try {
      final token = _registeredToken ?? await messaging.token();
      if (token != null) await _devices.unregister(token);
    } on Exception catch (e, stack) {
      _log('server unregister failed', e, stack);
    } finally {
      _registeredToken = null;
      try {
        await messaging.deleteToken();
      } on Exception catch (e, stack) {
        _log('deleting the FCM token failed', e, stack);
      }
    }
  }

  /// Stops listening. For tests and provider disposal.
  Future<void> dispose() async {
    await _tokenRefreshes?.cancel();
    await _tapData?.cancel();
    // Not awaited: a single-subscription controller that nobody listened to
    // only completes `close` once it is listened to.
    unawaited(_taps.close());
  }

  Future<void> _register(String token) async {
    final response = await _devices.register(
      token: token,
      platform: _platform,
      locale: _locale(),
    );
    if (response.success) {
      _registeredToken = token;
    } else {
      _log('server refused the token: ${response.statusCode} ${response.key}');
    }
  }

  void _onTapData(Map<String, Object?> data) =>
      _taps.add(PushTap.fromData(data));

  void _log(String message, [Object? error, StackTrace? stack]) =>
      developer.log(message, name: 'push', error: error, stackTrace: stack);
}
