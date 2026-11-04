import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/session.dart';
import 'package:letslive_mobile/features/auth/presentation/account_setup_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/splash_screen.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

import '../../features/auth/auth_test_app.dart';

Session sessionWith(String username) => Session(
  MeUser(
    id: 'u-1',
    username: username,
    email: 'user@example.com',
    status: 'normal',
    authProvider: 'google',
    createdAt: DateTime.utc(2026, 11, 4),
    followerCount: 0,
    streamApiKey: 'k',
  ),
);

void main() {
  group('authRedirect', () {
    const loading = AsyncLoading<Session?>();
    const loggedOut = AsyncData<Session?>(null);
    final ready = AsyncData<Session?>(sessionWith('abcdef'));
    final needsSetup = AsyncData<Session?>(sessionWith(''));

    test('waits on the splash while the session loads or failed', () {
      expect(authRedirect(loading, '/'), AppRoutes.splash);
      expect(authRedirect(loading, AppRoutes.login), AppRoutes.splash);
      expect(authRedirect(loading, AppRoutes.splash), isNull);
      final failed = AsyncError<Session?>(Exception('down'), StackTrace.empty);
      expect(authRedirect(failed, '/'), AppRoutes.splash);
    });

    test('logged out may only see login and signup', () {
      expect(authRedirect(loggedOut, '/'), AppRoutes.login);
      expect(authRedirect(loggedOut, AppRoutes.accountSetup), AppRoutes.login);
      expect(authRedirect(loggedOut, AppRoutes.splash), AppRoutes.login);
      expect(authRedirect(loggedOut, AppRoutes.login), isNull);
      expect(authRedirect(loggedOut, AppRoutes.signup), isNull);
    });

    test('missing username forces account setup everywhere', () {
      for (final location in [
        '/',
        AppRoutes.login,
        AppRoutes.signup,
        AppRoutes.splash,
        '/users/someone',
      ]) {
        expect(
          authRedirect(needsSetup, location),
          AppRoutes.accountSetup,
          reason: location,
        );
      }
      expect(authRedirect(needsSetup, AppRoutes.accountSetup), isNull);
    });

    test('a complete session leaves auth routes for home', () {
      for (final location in [
        AppRoutes.login,
        AppRoutes.signup,
        AppRoutes.accountSetup,
        AppRoutes.splash,
      ]) {
        expect(authRedirect(ready, location), AppRoutes.home, reason: location);
      }
      expect(authRedirect(ready, '/'), isNull);
    });
  });

  testWidgets(
    'user with null username is redirected to account-setup from any route',
    (tester) async {
      // The backend sends an empty username until a Google signup picks one.
      final app = await pumpTestApp(
        tester,
        server: FakeAuthServer(username: ''),
        signedIn: true,
      );
      expect(find.byType(AccountSetupScreen), findsOneWidget);

      final router = app.container.read(appRouterProvider);
      for (final location in ['/', AppRoutes.login, AppRoutes.signup]) {
        router.go(location);
        await tester.pumpAndSettle();
        expect(
          find.byType(AccountSetupScreen),
          findsOneWidget,
          reason: location,
        );
      }
    },
  );

  testWidgets('shows the splash, not login, while the session loads', (
    tester,
  ) async {
    final server = FakeAuthServer()..holdMe = Completer<void>();
    await pumpTestApp(tester, server: server, signedIn: true, settle: false);

    await tester.pump();
    expect(find.byType(SplashScreen), findsOneWidget);
    expect(find.byType(LoginScreen), findsNothing);

    server.holdMe?.complete();
    await tester.pumpAndSettle();
    expect(find.text('LetsLive (dev)'), findsOneWidget);
  });

  testWidgets('a failed session load shows an error with retry', (
    tester,
  ) async {
    final server = FakeAuthServer()..meStatus = 503;
    final app = await pumpTestApp(tester, server: server, signedIn: true);

    expect(find.byType(SplashScreen), findsOneWidget);
    expect(find.text('Oops!'), findsOneWidget);

    server.meStatus = 200;
    await tester.tap(find.byKey(const ValueKey('splash-retry')));
    await tester.pumpAndSettle();

    expect(find.text('LetsLive (dev)'), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.userMe), hasLength(2));
  });

  testWidgets('an expired session returns to login once', (tester) async {
    final app = await pumpTestApp(tester, signedIn: true);
    expect(find.text('LetsLive (dev)'), findsOneWidget);

    // The server rejects the refresh token: ApiClient reports sessionExpired.
    // Widget tests run in fake time: start the call, then pump it through.
    final refreshed = app.container.read(apiClientProvider).refreshSession();
    await tester.pumpAndSettle();
    expect(await refreshed, isFalse);

    expect(find.byType(LoginScreen), findsOneWidget);
    expect(await app.hasSessionCookie(), isFalse);
    expect(app.requestsTo(ApiEndpoints.logout), isEmpty);
    expect(app.requestsTo(ApiEndpoints.refreshToken), hasLength(1));
  });

  testWidgets('logout returns to login and clears the session', (tester) async {
    final app = await pumpTestApp(tester, signedIn: true);
    var hookRan = false;
    app.container
        .read(sessionLifecycleProvider)
        .addLogoutHook(() async => hookRan = true);

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;

    expect(hookRan, isTrue);
    expect(find.byType(LoginScreen), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.logout), hasLength(1));
    expect(await app.hasSessionCookie(), isFalse);
    expect(app.container.read(sessionProvider).value, isNull);
  });
}
