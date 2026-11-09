import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';

import '../auth/auth_test_app.dart';
import '../home/home_fixtures.dart';
import 'profile_test_app.dart';

/// Binding mobile state rule: log in as A, log out, log in as B -> fresh
/// requests and nothing of A's profile data (follow state, own profile,
/// videos).
void main() {
  testWidgets('the next account never sees the previous account\'s profiles', (
    tester,
  ) async {
    final server = FakeProfileServer()..channelVods = vodsOf(2);
    final app = await pumpRoute(tester, server, AppRoutes.channel('u-2'));

    // Account A follows u-2 and looks at their videos.
    await tester.tap(find.byKey(const ValueKey('follow-button')));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(Tab, 'Videos'));
    await tester.pumpAndSettle();
    expect(find.text('Unfollow'), findsOneWidget);
    expect(find.text('8 followers'), findsOneWidget);
    expect(find.text('Video 0'), findsOneWidget);
    expect(app.requestsTo('${ApiEndpoints.user}/u-2'), hasLength(1));

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);

    // The server now answers for account B, who follows nobody.
    server
      ..meId = 'u-9'
      ..username = 'bravo'
      ..channelVods = [vodJson(id: 'b-1', userId: 'u-2', title: 'B video')];
    server.users['u-2'] = publicUserJson(id: 'u-2', username: 'other');
    server.users['u-9'] = publicUserJson(id: 'u-9', username: 'bravo');
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;

    unawaited(
      app.container.read(appRouterProvider).push(AppRoutes.channel('u-2')),
    );
    await tester.pumpAndSettle();
    expect(app.requestsTo('${ApiEndpoints.user}/u-2'), hasLength(2));
    expect(find.text('Follow'), findsOneWidget);
    expect(find.text('Unfollow'), findsNothing);
    expect(find.text('7 followers'), findsOneWidget);

    await tester.tap(find.widgetWithText(Tab, 'Videos'));
    await tester.pumpAndSettle();
    expect(find.text('B video'), findsOneWidget);
    expect(find.text('Video 0'), findsNothing);

    await tester.binding.handlePopRoute(); // System back: no app bar here.
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Profile'));
    await tester.pumpAndSettle();
    expect(app.requestsTo('${ApiEndpoints.user}/u-9'), hasLength(1));
    expect(find.text('bravo'), findsOneWidget);
  });
}
