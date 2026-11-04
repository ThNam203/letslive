import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/data/google_sign_in_service.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:mocktail/mocktail.dart';

import 'auth_test_app.dart';

class MockGoogleSignIn extends Mock implements GoogleSignIn {}

class MockGoogleAccount extends Mock implements GoogleSignInAccount {}

Future<void> fillLogin(WidgetTester tester, String password) async {
  await enter(tester, 'email', 'user@example.com');
  await enter(tester, 'password', password);
}

void main() {
  testWidgets('logged-out user lands on the login screen', (tester) async {
    await pumpTestApp(tester);

    expect(find.byType(LoginScreen), findsOneWidget);
    expect(find.text('Welcome back!'), findsOneWidget);
  });

  testWidgets('login stores session and navigates home', (tester) async {
    final app = await pumpTestApp(tester);

    await fillLogin(tester, testPassword);
    await tester.tap(find.byKey(const ValueKey('login-submit')));
    await tester.pumpAndSettle();

    expect(find.byType(HomeScreen), findsOneWidget);
    expect(await app.hasSessionCookie(), isTrue);
    final session = app.container.read(sessionProvider).value;
    expect(session?.user.username, 'abcdef');
    final login = app.requestsTo(ApiEndpoints.login).single;
    expect(login.headers['X-Firebase-AppCheck'], isNotEmpty);
  });

  testWidgets('shows progress and blocks double submit while logging in', (
    tester,
  ) async {
    final server = FakeAuthServer()..holdLogin = Completer<void>();
    final app = await pumpTestApp(tester, server: server);

    await fillLogin(tester, testPassword);
    await tester.tap(find.byKey(const ValueKey('login-submit')));
    await tester.pump();

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    await tester.tap(
      find.byKey(const ValueKey('login-submit')),
      warnIfMissed: false,
    );
    await tester.pump();

    server.holdLogin?.complete();
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.login), hasLength(1));
  });

  testWidgets('wrong password shows server message with retry', (tester) async {
    final app = await pumpTestApp(tester);

    await fillLogin(tester, 'wrong-password');
    await tester.tap(find.byKey(const ValueKey('login-submit')));
    await tester.pumpAndSettle();

    // The localized message for the server key.
    expect(find.text('Username or password incorrect.'), findsOneWidget);
    expect(find.byType(LoginScreen), findsOneWidget);
    expect(await app.hasSessionCookie(), isFalse);

    // Retry: fix the password and submit again with a fresh App Check token.
    await enter(tester, 'password', testPassword);
    await tester.tap(find.byKey(const ValueKey('login-submit')));
    await tester.pumpAndSettle();

    expect(find.byType(HomeScreen), findsOneWidget);
    final attempts = app.requestsTo(ApiEndpoints.login);
    expect(attempts, hasLength(2));
    expect(
      attempts[0].headers['X-Firebase-AppCheck'],
      isNot(attempts[1].headers['X-Firebase-AppCheck']),
    );
  });

  testWidgets('invalid input is rejected before any request', (tester) async {
    final app = await pumpTestApp(tester);

    await enter(tester, 'email', 'not-an-email');
    await enter(tester, 'password', 'short');
    await tester.tap(find.byKey(const ValueKey('login-submit')));
    await tester.pumpAndSettle();

    expect(find.text('Email is invalid'), findsOneWidget);
    expect(find.text('Password must be at least 8 characters'), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.login), isEmpty);
  });

  testWidgets('Google button is hidden without a Google client id', (
    tester,
  ) async {
    await pumpTestApp(tester);

    expect(find.byKey(const ValueKey('google-sign-in')), findsNothing);
  });

  group('Google sign-in (fake at the google_sign_in boundary)', () {
    late MockGoogleSignIn signIn;

    setUp(() {
      signIn = MockGoogleSignIn();
      when(
        () => signIn.initialize(
          clientId: any(named: 'clientId'),
          serverClientId: any(named: 'serverClientId'),
        ),
      ).thenAnswer((_) async {});
    });

    GoogleSignInService service() => GoogleSignInService(
      serverClientId: 'web-client.apps.googleusercontent.com',
      signIn: signIn,
    );

    testWidgets('google button sends the native id token and signs in', (
      tester,
    ) async {
      final account = MockGoogleAccount();
      when(() => account.authentication).thenReturn(
        const GoogleSignInAuthentication(idToken: 'native-id-token'),
      );
      when(() => signIn.authenticate()).thenAnswer((_) async => account);
      final app = await pumpTestApp(tester, google: service());

      await tester.tap(find.byKey(const ValueKey('google-sign-in')));
      await tester.pumpAndSettle();

      final request = app.requestsTo(ApiEndpoints.googleMobile).single;
      expect(request.data, {'idToken': 'native-id-token'});
      expect(find.byType(HomeScreen), findsOneWidget);
      verify(
        () => signIn.initialize(
          serverClientId: 'web-client.apps.googleusercontent.com',
        ),
      ).called(1);
    });

    testWidgets('cancelling the Google sheet sends nothing', (tester) async {
      when(() => signIn.authenticate()).thenThrow(
        const GoogleSignInException(code: GoogleSignInExceptionCode.canceled),
      );
      final app = await pumpTestApp(tester, google: service());

      await tester.tap(find.byKey(const ValueKey('google-sign-in')));
      await tester.pumpAndSettle();

      expect(app.requestsTo(ApiEndpoints.googleMobile), isEmpty);
      expect(find.byType(LoginScreen), findsOneWidget);
    });

    testWidgets('a Google SDK error shows a generic message, not the '
        'SDK description', (tester) async {
      when(() => signIn.authenticate()).thenThrow(
        const GoogleSignInException(
          code: GoogleSignInExceptionCode.clientConfigurationError,
          description: 'SHA-1 not registered',
        ),
      );
      await pumpTestApp(tester, google: service());

      await tester.tap(find.byKey(const ValueKey('google-sign-in')));
      await tester.pumpAndSettle();

      expect(find.textContaining('SHA-1 not registered'), findsNothing);
      expect(
        find.text('Something went wrong. Please try again.'),
        findsOneWidget,
      );
      expect(find.byType(LoginScreen), findsOneWidget);
    });

    testWidgets('a non-Google error (e.g. a platform failure) shows a '
        'generic message', (tester) async {
      when(() => signIn.authenticate())
          .thenThrow(PlatformException(code: 'channel-error'));
      await pumpTestApp(tester, google: service());

      await tester.tap(find.byKey(const ValueKey('google-sign-in')));
      await tester.pumpAndSettle();

      expect(
        find.text('Something went wrong. Please try again.'),
        findsOneWidget,
      );
      // The button is usable again.
      final button = tester.widget<ButtonStyleButton>(
        find.byKey(const ValueKey('google-sign-in')),
      );
      expect(button.onPressed, isNotNull);
    });
  });
}
