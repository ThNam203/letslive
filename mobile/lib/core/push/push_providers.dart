/// Riverpod wiring of push notifications: the service, its Firebase
/// boundary, and the bridge to the session and the router.
library;

import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/push/device_repository.dart';
import 'package:letslive_mobile/core/push/push_messaging.dart';
import 'package:letslive_mobile/core/push/push_service.dart';
import 'package:letslive_mobile/core/push/push_tap.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';

/// FCM, or null when Firebase is not configured (dev without `FIREBASE_*`),
/// which makes [PushService] a no-op. Tests override it with a fake.
final pushMessagingProvider = Provider<PushMessaging?>(
  (ref) => ref.watch(appConfigProvider).firebase == null
      ? null
      : const FirebasePushMessaging(),
);

/// The device registration endpoints.
final deviceRepositoryProvider = Provider<DeviceRepository>(
  (ref) => DeviceRepository(ref.watch(apiClientProvider)),
);

/// The one [PushService]. Creating it registers its `unregister` as a logout
/// hook, so the app reads it at boot (`main`, and [pushBridgeProvider]) before
/// anyone can log out.
final pushServiceProvider = Provider<PushService>((ref) {
  final service = PushService(
    messaging: ref.watch(pushMessagingProvider),
    devices: ref.watch(deviceRepositoryProvider),
    platform: defaultTargetPlatform == TargetPlatform.iOS ? 'ios' : 'android',
    // Read on each registration, so it is the language the app shows then.
    locale: () => ref.read(localeProvider).toLanguageTag(),
  );
  ref.read(sessionLifecycleProvider).addLogoutHook(service.unregister);
  ref.onDispose(() => unawaited(service.dispose()));
  return service;
});

/// Connects push to the rest of the app; `LetsLiveApp` watches it for the
/// whole run.
///
/// - Registers the device each time a user with a finished account signs in
///   (also at a cold start with a saved session).
/// - Re-sends the device's language when the user changes the app language.
/// - Opens the screen a tapped push names ([resolveDeepLink]). A tap that
///   launched the app waits for the session to load first; for a signed-out
///   user the tap is dropped (the router would only show the login).
final pushBridgeProvider = Provider<void>((ref) {
  final push = ref.watch(pushServiceProvider);
  ref.listen(
    sessionProvider.select((s) {
      final session = s.value;
      return session == null || session.needsAccountSetup
          ? null
          : session.user.id;
    }),
    (_, userId) {
      if (userId != null) unawaited(push.registerForUser());
    },
    fireImmediately: true,
  );
  ref.listen(localeProvider, (_, _) => unawaited(push.resendLocale()));
  final taps = push.taps.listen((tap) => unawaited(_openTap(ref, tap)));
  ref.onDispose(() => unawaited(taps.cancel()));
});

Future<void> _openTap(Ref ref, PushTap tap) async {
  final session = await ref
      .read(sessionProvider.future)
      .then<Object?>((s) => s, onError: (Object _) => null);
  if (session == null || !ref.mounted) return;
  final route = resolveDeepLink(tap);
  final router = ref.read(appRouterProvider);
  // Pushed over the current screen so back returns to it; a tap for the
  // screen already on top does not stack a second copy.
  if (router.state.uri.path == route) return;
  unawaited(router.push<void>(route));
}
