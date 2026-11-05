/// The Firebase Cloud Messaging boundary: everything `PushService` needs from
/// `firebase_messaging`, behind a small interface so tests use a fake.
library;

import 'package:firebase_messaging/firebase_messaging.dart';

/// What the app uses of FCM.
abstract interface class PushMessaging {
  /// Makes a push that arrives while the app is open show nothing. The app
  /// is connected to the realtime socket then (and the server skips the
  /// push for connected users), so the in-app UI already shows it.
  Future<void> silenceForeground();

  /// Asks the user to allow notifications (the Android 13+ and iOS prompt).
  /// True when allowed.
  Future<bool> requestPermission();

  /// This install's FCM token; null when FCM has none (yet).
  Future<String?> token();

  /// New tokens after FCM rotated the old one.
  Stream<String> get tokenRefreshes;

  /// Drops this install's token, so a logged-out device gets no more pushes.
  Future<void> deleteToken();

  /// The data of the push whose tap launched the app from a closed state.
  Future<Map<String, Object?>?> launchTapData();

  /// The data of each push tapped while the app was in the background.
  Stream<Map<String, Object?>> get tapData;
}

/// [PushMessaging] over `FirebaseMessaging.instance`. Only built when
/// Firebase is configured (`FIREBASE_*` keys), after `Firebase.initializeApp`.
class FirebasePushMessaging implements PushMessaging {
  const FirebasePushMessaging();

  FirebaseMessaging get _fcm => FirebaseMessaging.instance;

  @override
  Future<void> silenceForeground() =>
      _fcm.setForegroundNotificationPresentationOptions();

  @override
  Future<bool> requestPermission() async {
    final settings = await _fcm.requestPermission();
    return switch (settings.authorizationStatus) {
      AuthorizationStatus.authorized || AuthorizationStatus.provisional => true,
      AuthorizationStatus.denied ||
      AuthorizationStatus.deniedPermanently ||
      AuthorizationStatus.notDetermined => false,
    };
  }

  @override
  Future<String?> token() => _fcm.getToken();

  @override
  Stream<String> get tokenRefreshes => _fcm.onTokenRefresh;

  @override
  Future<void> deleteToken() => _fcm.deleteToken();

  @override
  Future<Map<String, Object?>?> launchTapData() async {
    final message = await _fcm.getInitialMessage();
    return message == null ? null : Map<String, Object?>.of(message.data);
  }

  @override
  Stream<Map<String, Object?>> get tapData => FirebaseMessaging
      .onMessageOpenedApp
      .map((message) => Map<String, Object?>.of(message.data));
}
