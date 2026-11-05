/// Auth and session for the app. Mirrors web `lib/api/auth.ts` and the
/// `(auth)` pages.
///
/// How it fits together:
/// - The session is cookie based. Login, signup and Google sign-in make the
///   server set ACCESS_TOKEN/REFRESH_TOKEN cookies in [ApiClient]'s secure
///   cookie jar; nothing else stores tokens.
/// - [sessionProvider] is the single source of truth: it loads
///   `GET /user/me` at startup (401 = logged out) and after every sign-in.
///   The router (`core/router/app_router.dart`) redirects from it:
///   logged out -> `/login`, no username yet (Google signup) ->
///   `/account-setup` from any route, signed in -> the app.
/// - Login and verify-email carry a bot proof ([botProofProvider]): a fresh
///   Firebase App Check token per request, or the Turnstile test token in
///   dev without Firebase.
/// - Logout and session expiry (the server rejected the refresh token, see
///   [ApiClient.sessionExpired]) both go through [SessionLifecycle]: logout
///   hooks (e.g. push unregister), `DELETE /auth/logout` (logout only),
///   clear the cookie jar, restart the realtime socket, then the session
///   becomes null and the router shows `/login` once.
library;

import 'dart:async';
import 'dart:developer' as developer;

import 'package:firebase_app_check/firebase_app_check.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';
import 'package:letslive_mobile/features/auth/data/google_sign_in_service.dart';
import 'package:letslive_mobile/features/auth/domain/session.dart';
import 'package:letslive_mobile/features/auth/domain/session_lifecycle.dart';
import 'package:letslive_mobile/features/user/data/me_repository.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// App Check when Firebase is configured, else the dev Turnstile token
/// (`AppConfig` guarantees one of the two).
final botProofProvider = Provider<BotProof>((ref) {
  final config = ref.watch(appConfigProvider);
  if (config.firebase != null) {
    return BotProof.appCheck(FirebaseAppCheck.instance.getLimitedUseToken);
  }
  final token = config.turnstileDevToken;
  if (token == null) {
    throw StateError('AppConfig has neither Firebase nor a Turnstile token.');
  }
  return BotProof.turnstile(token);
});

/// Auth endpoints.
final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => AuthRepository(
    client: ref.watch(apiClientProvider),
    botProof: ref.watch(botProofProvider),
  ),
);

/// `/user/me` reads and the account-setup update.
final meRepositoryProvider = Provider<MeRepository>(
  (ref) => MeRepository(ref.watch(apiClientProvider)),
);

/// Native Google sign-in; null (button hidden) without a Google client id.
final googleSignInServiceProvider = Provider<GoogleSignInService?>((ref) {
  final google = ref.watch(appConfigProvider).googleSignIn;
  if (google == null) return null;
  return GoogleSignInService(
    serverClientId: google.serverClientId,
    iosClientId: google.iosClientId,
  );
});

/// Logout hooks and the end-of-session sequence. Features register their
/// cleanup with [SessionLifecycle.addLogoutHook].
///
/// The registered hooks live on this one instance, so it must never be
/// rebuilt: its inputs are read (not watched) and are fixed for the app's
/// lifetime (config and overrides set at boot).
final sessionLifecycleProvider = Provider<SessionLifecycle>((ref) {
  final lifecycle = SessionLifecycle(
    auth: ref.read(authRepositoryProvider),
    cookieJar: ref.read(cookieJarProvider),
    restartRealtime: () => ref.read(realtimeManagerProvider).restart(),
  );
  // Forget the Google account too, so the next Google sign-in asks again.
  final google = ref.read(googleSignInServiceProvider);
  if (google != null) lifecycle.addLogoutHook(google.signOut);
  return lifecycle;
});

/// `GET /user/me` answered with something other than a user or a 401.
class SessionLoadException implements Exception {
  const SessionLoadException(this.response);

  final ApiResponse<Object?> response;

  @override
  String toString() =>
      'SessionLoadException(${response.statusCode} ${response.key})';
}

/// The current session; null when logged out. Errors (server or network
/// trouble while loading) are shown by the splash screen with a retry.
final sessionProvider = AsyncNotifierProvider<SessionNotifier, Session?>(
  SessionNotifier.new,
  // The splash offers a manual retry; no automatic background retries.
  retry: (_, _) => null,
);

/// Loads the session and runs sign-in, account setup and sign-out.
class SessionNotifier extends AsyncNotifier<Session?> {
  @override
  Future<Session?> build() async {
    final expired = ref
        .watch(apiClientProvider)
        .sessionExpired
        .listen((_) => unawaited(_endSession(notifyServer: false)));
    ref.onDispose(expired.cancel);
    return _load();
  }

  Future<Session?> _load() async {
    final res = await ref.read(meRepositoryProvider).fetch();
    final user = res.data;
    if (res.success && user != null) return Session(user);
    if (res.statusCode == 401) return null;
    throw SessionLoadException(res);
  }

  /// Email and password login.
  Future<ApiResponse<void>> login(String email, String password) =>
      _signIn(ref.read(authRepositoryProvider).login(email, password));

  /// Creates the account with the emailed code and signs in.
  Future<ApiResponse<void>> signup({
    required String email,
    required String username,
    required String password,
    required String otpCode,
  }) => _signIn(
    ref
        .read(authRepositoryProvider)
        .signup(
          email: email,
          username: username,
          password: password,
          otpCode: otpCode,
        ),
  );

  /// Signs in with a Google id token from [GoogleSignInService].
  Future<ApiResponse<void>> loginWithGoogle(String idToken) =>
      _signIn(ref.read(authRepositoryProvider).loginWithGoogle(idToken));

  /// Sets the username of a session that has none yet.
  Future<ApiResponse<void>> completeAccountSetup(String username) =>
      updateProfile(username: username);

  /// Saves the given profile fields (`PUT /user/me`, see
  /// `MeRepository.update`) and, on success, replaces the session user with
  /// the server's answer, so every screen showing the own profile updates
  /// at once. Account setup uses it too.
  Future<ApiResponse<void>> updateProfile({
    String? username,
    String? bio,
    Map<String, String>? socialMediaLinks,
    String? status,
    String? phoneNumber,
    String? locale,
  }) async {
    final res = await ref
        .read(meRepositoryProvider)
        .update(
          username: username,
          bio: bio,
          socialMediaLinks: socialMediaLinks,
          status: status,
          phoneNumber: phoneNumber,
          locale: locale,
        );
    final user = res.data;
    if (res.success && user != null) state = AsyncData(Session(user));
    return res;
  }

  /// Uploads a new profile picture; on success the session user gets the
  /// new URL.
  Future<ApiResponse<String>> uploadProfilePicture(
    PickedImage image, {
    void Function(int sent, int total)? onProgress,
  }) async {
    final res = await ref
        .read(meRepositoryProvider)
        .uploadProfilePicture(image, onProgress: onProgress);
    final url = res.data;
    final current = state.value?.user;
    if (res.success && url != null && current != null) {
      state = AsyncData(Session(current.withPictures(profilePicture: url)));
    }
    return res;
  }

  /// Uploads a new profile banner; on success the session user gets the new
  /// URL.
  Future<ApiResponse<String>> uploadBackgroundPicture(
    PickedImage image, {
    void Function(int sent, int total)? onProgress,
  }) async {
    final res = await ref
        .read(meRepositoryProvider)
        .uploadBackgroundPicture(image, onProgress: onProgress);
    final url = res.data;
    final current = state.value?.user;
    if (res.success && url != null && current != null) {
      state = AsyncData(Session(current.withPictures(backgroundPicture: url)));
    }
    return res;
  }

  /// User-initiated logout.
  Future<void> logout() => _endSession(notifyServer: true);

  Future<ApiResponse<void>> _signIn(Future<ApiResponse<void>> call) async {
    final res = await call;
    if (!res.success) return res;
    ref.read(sessionLifecycleProvider).signedIn();
    // No loading state here: the screen shows its own progress, and the
    // router must not flash the splash.
    state = await AsyncValue.guard(_load);
    return res;
  }

  // The app is signed out even when cleanup fails (e.g. the cookie wipe), so
  // the user is never stuck in a half-ended session; the failure is logged.
  // Not rethrown: the expiry path runs unawaited and has no caller to tell.
  // Object, like the SessionLifecycle hook loop: any failure is logged.
  Future<void> _endSession({required bool notifyServer}) async {
    try {
      await ref
          .read(sessionLifecycleProvider)
          .endSession(notifyServer: notifyServer);
    } on Object catch (e, stack) {
      developer.log(
        'Ending the session failed',
        name: 'auth',
        error: e,
        stackTrace: stack,
      );
    } finally {
      state = const AsyncData(null);
    }
  }
}
