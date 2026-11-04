import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/signup_screen.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';

import 'auth_test_app.dart';

Future<TestApp> openSignup(WidgetTester tester) async {
  final app = await pumpTestApp(tester);
  await tester.tap(find.byKey(const ValueKey('go-signup')));
  await tester.pumpAndSettle();
  expect(find.byType(SignupScreen), findsOneWidget);
  return app;
}

Future<void> fillSignup(WidgetTester tester) async {
  await enter(tester, 'email', 'new@example.com');
  await enter(tester, 'username', 'new_user');
  await enter(tester, 'password', testPassword);
  await enter(tester, 'confirm-password', testPassword);
}

int? maxLengthOf(WidgetTester tester, String key) =>
    tester.widget<TextField>(find.byKey(ValueKey(key))).maxLength;

void main() {
  testWidgets('signup validates same field limits as web', (tester) async {
    final app = await openSignup(tester);

    // Input caps mirror the web inputs' maxLength.
    expect(maxLengthOf(tester, 'email'), 320);
    expect(maxLengthOf(tester, 'username'), 30);
    expect(maxLengthOf(tester, 'password'), 72);
    expect(maxLengthOf(tester, 'confirm-password'), 72);

    await enter(tester, 'email', 'new@example.com');
    await enter(tester, 'username', 'short');
    await enter(tester, 'password', 'alllowercase1!');
    await enter(tester, 'confirm-password', 'different');
    await tester.tap(find.byKey(const ValueKey('signup-submit')));
    await tester.pumpAndSettle();

    expect(find.text('Username must be >= 6 characters'), findsOneWidget);
    expect(
      find.text('Password must contain at least one uppercase letter'),
      findsOneWidget,
    );
    expect(find.text('Passwords do not match'), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.verifyEmail), isEmpty);
  });

  testWidgets('verification code flow creates the account and goes home', (
    tester,
  ) async {
    final app = await openSignup(tester);
    await fillSignup(tester);

    await tester.tap(find.byKey(const ValueKey('signup-submit')));
    await tester.pumpAndSettle();

    final verify = app.requestsTo(ApiEndpoints.verifyEmail).single;
    expect(verify.headers['X-Firebase-AppCheck'], isNotEmpty);
    expect(find.text('Enter Verification Code'), findsOneWidget);

    // A wrong code shows the server message and keeps the dialog open.
    await enter(tester, 'otp', '000000');
    await tester.tap(find.byKey(const ValueKey('otp-submit')));
    await tester.pumpAndSettle();
    expect(find.text('OTP code not found.'), findsOneWidget);

    await enter(tester, 'otp', testOtp);
    await tester.tap(find.byKey(const ValueKey('otp-submit')));
    await tester.pumpAndSettle();

    expect(find.byType(HomeScreen), findsOneWidget);
    expect(await app.hasSessionCookie(), isTrue);
    final signup = app.requestsTo(ApiEndpoints.signup).last;
    expect(signup.data, {
      'email': 'new@example.com',
      'username': 'new_user',
      'password': testPassword,
      'otpCode': testOtp,
    });
  });

  testWidgets('a failed verification request shows the error and retries', (
    tester,
  ) async {
    final app = await openSignup(tester);
    final handle = app.server.handle;
    var failNext = true;
    app.adapter.handler = (o) {
      if (o.path == ApiEndpoints.verifyEmail && failNext) {
        failNext = false;
        throw StateError('connection reset');
      }
      return handle(o);
    };
    await fillSignup(tester);

    await tester.tap(find.byKey(const ValueKey('signup-submit')));
    await tester.pumpAndSettle();
    expect(
      find.text('Failed to fetch data, please try again.'),
      findsOneWidget,
    );

    await tester.tap(find.byKey(const ValueKey('signup-submit')));
    await tester.pumpAndSettle();
    expect(find.text('Enter Verification Code'), findsOneWidget);
  });

  testWidgets('shows progress and blocks double submit while sending the '
      'code and while creating the account', (tester) async {
    final app = await openSignup(tester);
    app.server
      ..holdVerifyEmail = Completer<void>()
      ..holdSignup = Completer<void>();
    await fillSignup(tester);

    await tester.tap(find.byKey(const ValueKey('signup-submit')));
    await tester.pump();
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    expect(_onPressed(tester, 'signup-submit'), isNull);
    await tester.tap(
      find.byKey(const ValueKey('signup-submit')),
      warnIfMissed: false,
    );
    app.server.holdVerifyEmail?.complete();
    await tester.pumpAndSettle();
    expect(app.requestsTo(ApiEndpoints.verifyEmail), hasLength(1));

    await enter(tester, 'otp', testOtp);
    await tester.tap(find.byKey(const ValueKey('otp-submit')));
    await tester.pump();
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    expect(_onPressed(tester, 'otp-submit'), isNull);
    expect(
      tester.widget<TextField>(find.byKey(const ValueKey('otp'))).enabled,
      isFalse,
    );

    app.server.holdSignup?.complete();
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.signup), hasLength(1));
  });

  testWidgets('login link goes back to login', (tester) async {
    await openSignup(tester);

    await tester.tap(find.byKey(const ValueKey('go-login')));
    await tester.pumpAndSettle();

    expect(find.byType(LoginScreen), findsOneWidget);
  });
}

/// The tap handler of the button under [key] (null while disabled). The key
/// may sit on a wrapper such as `SubmitButton`.
VoidCallback? _onPressed(WidgetTester tester, String key) => tester
    .widget<ButtonStyleButton>(
      find.descendant(
        of: find.byKey(ValueKey(key)),
        matching: find.bySubtype<ButtonStyleButton>(),
        matchRoot: true,
      ),
    )
    .onPressed;
