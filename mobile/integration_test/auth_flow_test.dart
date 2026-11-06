import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/main.dart' as app;

/// Credentials of an existing dev-stack user (the app cannot read signup
/// codes from the database on a device). See mobile/README.md.
const _email = String.fromEnvironment('AUTH_FLOW_EMAIL');
const _password = String.fromEnvironment('AUTH_FLOW_PASSWORD');

const _timeout = Duration(seconds: 20);

/// Pumps until [finder] matches; the splash spinner never lets
/// `pumpAndSettle` finish, so this polls instead.
Future<void> waitFor(WidgetTester tester, Finder finder) async {
  final deadline = DateTime.now().add(_timeout);
  while (finder.evaluate().isEmpty) {
    if (DateTime.now().isAfter(deadline)) {
      throw TestFailure('Timed out waiting for $finder');
    }
    await tester.pump(const Duration(milliseconds: 200));
  }
}

ProviderContainer containerOf(WidgetTester tester) =>
    ProviderScope.containerOf(tester.element(find.byType(LetsLiveApp)));

Future<void> logIn(WidgetTester tester, String password) async {
  await tester.enterText(find.byKey(const ValueKey('email')), _email);
  await tester.enterText(find.byKey(const ValueKey('password')), password);
  await tester.tap(find.byKey(const ValueKey('login-submit')));
}

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('login, wrong password, logout against the dev stack', (
    tester,
  ) async {
    expect(_email, isNotEmpty, reason: 'pass --dart-define=AUTH_FLOW_EMAIL');
    expect(_password, isNotEmpty, reason: 'pass AUTH_FLOW_PASSWORD');

    await app.main();
    await waitFor(
      tester,
      find.byWidgetPredicate((w) => w is LoginScreen || w is HomeScreen),
    );
    // A previous run may have left a session in the keychain.
    if (find.byType(HomeScreen).evaluate().isNotEmpty) {
      await containerOf(tester).read(sessionProvider.notifier).logout();
      await waitFor(tester, find.byType(LoginScreen));
    }

    await logIn(tester, 'Wrong-Pass1!');
    await waitFor(tester, find.text('Username or password incorrect.'));

    await logIn(tester, _password);
    await waitFor(tester, find.byType(HomeScreen));
    final session = containerOf(tester).read(sessionProvider).value;
    expect(session?.user.email, _email);

    await containerOf(tester).read(sessionProvider.notifier).logout();
    await waitFor(tester, find.byType(LoginScreen));
    expect(containerOf(tester).read(sessionProvider).value, isNull);
  });
}
