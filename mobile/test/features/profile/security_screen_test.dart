import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/router/app_router.dart';

import '../auth/auth_test_app.dart';
import 'profile_test_app.dart';

final _submit = find.byKey(const ValueKey('password-submit'));

Future<void> _fill(
  WidgetTester tester, {
  String current = testPassword,
  String next = 'New-pass-1!',
  String? confirm,
}) async {
  await tester.enterText(field('current-password'), current);
  await tester.enterText(field('new-password'), next);
  await tester.enterText(field('confirm-password'), confirm ?? next);
  await tester.pump();
}

/// `/settings/security`: the email and the change-password form.
void main() {
  testWidgets('shows the account email and the form', (tester) async {
    await pumpRoute(tester, FakeProfileServer(), AppRoutes.settingsSecurity);

    expect(find.text('user@example.com'), findsOneWidget);
    expect(find.text('Current Password'), findsOneWidget);
    expect(find.text('New Password'), findsOneWidget);
    expect(find.text('Confirm New Password'), findsOneWidget);
  });

  testWidgets('passwords shorter than 8 are rejected with web\'s message', (
    tester,
  ) async {
    final server = FakeProfileServer();
    await pumpRoute(tester, server, AppRoutes.settingsSecurity);

    await _fill(tester, current: 'short', next: 'tiny', confirm: 'tiny');
    await tester.tap(_submit);
    await tester.pumpAndSettle();

    expect(
      find.text('Password must be at least 8 characters'),
      findsNWidgets(3),
    );
    expect(server.passwordBodies, isEmpty);
  });

  testWidgets('a new password equal to the current one is rejected', (
    tester,
  ) async {
    final server = FakeProfileServer();
    await pumpRoute(tester, server, AppRoutes.settingsSecurity);

    await _fill(tester, next: testPassword);
    await tester.tap(_submit);
    await tester.pumpAndSettle();

    expect(
      find.text('New password must be different from current password'),
      findsOneWidget,
    );
    expect(server.passwordBodies, isEmpty);
  });

  testWidgets('a confirmation that differs is rejected', (tester) async {
    final server = FakeProfileServer();
    await pumpRoute(tester, server, AppRoutes.settingsSecurity);

    await _fill(tester, confirm: 'Other-pass-1!');
    await tester.tap(_submit);
    await tester.pumpAndSettle();

    expect(find.text('Passwords do not match'), findsOneWidget);
    expect(server.passwordBodies, isEmpty);
  });

  testWidgets('a valid change sends both passwords and clears the form', (
    tester,
  ) async {
    final server = FakeProfileServer();
    await pumpRoute(tester, server, AppRoutes.settingsSecurity);

    await _fill(tester);
    await tester.tap(_submit);
    await tester.pumpAndSettle();

    expect(server.passwordBodies, [
      {'oldPassword': testPassword, 'newPassword': 'New-pass-1!'},
    ]);
    expect(find.text('Password updated successfully'), findsOneWidget);
    expect(
      tester.widget<TextField>(field('new-password')).controller!.text,
      isEmpty,
    );
  });

  testWidgets('the server refusing shows its message under the form', (
    tester,
  ) async {
    final server = FakeProfileServer();
    await pumpRoute(tester, server, AppRoutes.settingsSecurity);

    await _fill(tester, current: 'Wrong-pass-1!');
    await tester.tap(_submit);
    await tester.pumpAndSettle();

    expect(find.text('Old password does not match.'), findsOneWidget);
    expect(find.text('Password updated successfully'), findsNothing);
    // The typed text is kept so the user can fix it.
    expect(
      tester.widget<TextField>(field('new-password')).controller!.text,
      'New-pass-1!',
    );
  });

  testWidgets('an account without a password gets a note, no form', (
    tester,
  ) async {
    final server = FakeProfileServer()..meExtra = {'authProvider': 'google'};
    await pumpRoute(tester, server, AppRoutes.settingsSecurity);

    expect(field('current-password'), findsNothing);
    expect(
      find.text(
        'Your account is secured with a password from a third-party provider.',
      ),
      findsOneWidget,
    );
  });

  testWidgets('fits 320px at text scale 2', (tester) async {
    await pumpRoute(
      tester,
      FakeProfileServer(),
      AppRoutes.settingsSecurity,
      width: 320,
      height: 700,
      textScale: 2,
    );
    await _fill(tester, current: 'short');
    await tester.ensureVisible(_submit);
    await tester.pumpAndSettle();
    await tester.tap(_submit);
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('Password must be at least 8 characters'), findsWidgets);
  });
}
