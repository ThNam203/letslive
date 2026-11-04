import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/account_setup_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';

import 'auth_test_app.dart';

Future<TestApp> pumpSetup(WidgetTester tester) async {
  final app = await pumpTestApp(
    tester,
    server: FakeAuthServer(username: ''),
    signedIn: true,
  );
  expect(find.byType(AccountSetupScreen), findsOneWidget);
  return app;
}

void main() {
  testWidgets('a too short username is rejected locally', (tester) async {
    final app = await pumpSetup(tester);

    await enter(tester, 'username', '  abc  ');
    await tester.tap(find.byKey(const ValueKey('setup-submit')));
    await tester.pumpAndSettle();

    expect(find.text('Username must be >= 6 characters'), findsOneWidget);
    expect(app.adapter.requests.where((r) => r.method == 'PUT'), isEmpty);
  });

  testWidgets('a taken username shows the server message', (tester) async {
    await pumpSetup(tester);

    await enter(tester, 'username', 'taken_name');
    await tester.tap(find.byKey(const ValueKey('setup-submit')));
    await tester.pumpAndSettle();

    expect(find.byType(AccountSetupScreen), findsOneWidget);
    expect(find.text('Username already taken.'), findsOneWidget);
  });

  testWidgets('choosing a username unlocks the app', (tester) async {
    final app = await pumpSetup(tester);

    await enter(tester, 'username', '  chosen_name ');
    await tester.tap(find.byKey(const ValueKey('setup-submit')));
    await tester.pumpAndSettle();

    final put = app.adapter.requests.singleWhere((r) => r.method == 'PUT');
    expect(put.path, ApiEndpoints.userMe);
    expect(put.data, {'username': 'chosen_name'});
    expect(find.text('LetsLive (dev)'), findsOneWidget);
    expect(
      app.container.read(sessionProvider).value?.user.username,
      'chosen_name',
    );
  });

  testWidgets('the user can log out instead', (tester) async {
    await pumpSetup(tester);

    await tester.tap(find.byKey(const ValueKey('setup-logout')));
    await tester.pumpAndSettle();

    expect(find.byType(LoginScreen), findsOneWidget);
  });
}
