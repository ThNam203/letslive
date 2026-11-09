import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/watch/presentation/channel_screen.dart';

import '../auth/auth_test_app.dart';
import '../home/home_fixtures.dart';
import 'profile_test_app.dart';

/// Finder for something inside the section with [key].
Finder _inSection(String key, Finder matching) =>
    find.descendant(of: find.byKey(ValueKey(key)), matching: matching);

/// The signed-in user's profile on its People tab.
Future<TestApp> _pumpPeople(
  WidgetTester tester,
  FakeProfileServer server,
) async {
  final app = await pumpRoute(tester, server, AppRoutes.profile);
  await tester.tap(find.widgetWithText(Tab, 'People'));
  await tester.pumpAndSettle();
  return app;
}

void main() {
  testWidgets('People lists followed channels, then recommended ones not '
      'already followed', (tester) async {
    final server = FakeProfileServer()
      ..following = [publicUserJson(id: 'u-2', username: 'followed_one')]
      ..recommended = [
        publicUserJson(id: 'u-2', username: 'followed_one'),
        publicUserJson(id: 'u-3', username: 'suggested_one'),
      ];
    await _pumpPeople(tester, server);

    expect(
      _inSection('section-following', find.text('followed_one')),
      findsOne,
    );
    expect(_inSection('section-following', find.text('7 followers')), findsOne);
    expect(
      _inSection('section-recommended', find.text('followed_one')),
      findsNothing,
    );
    expect(
      _inSection('section-recommended', find.text('suggested_one')),
      findsOne,
    );
  });

  testWidgets('empty lists are left out, like web', (tester) async {
    await _pumpPeople(tester, FakeProfileServer());

    expect(find.byKey(const ValueKey('section-following')), findsNothing);
    expect(find.byKey(const ValueKey('section-recommended')), findsNothing);
  });

  testWidgets('a failed list shows its own retry', (tester) async {
    final server = FakeProfileServer()
      ..followingStatus = 503
      ..recommended = [publicUserJson(id: 'u-3', username: 'suggested_one')];
    final app = await _pumpPeople(tester, server);

    expect(_inSection('section-following', find.text('Try again')), findsOne);
    expect(find.text('suggested_one'), findsOne);

    server
      ..followingStatus = 200
      ..following = [publicUserJson(id: 'u-2', username: 'followed_one')];
    await tester.tap(_inSection('section-following', find.text('Try again')));
    await tester.pumpAndSettle();

    expect(find.text('followed_one'), findsOne);
    expect(app.requestsTo(ApiEndpoints.following), hasLength(2));
    expect(app.requestsTo(ApiEndpoints.recommendedChannels), hasLength(1));
  });

  testWidgets('pull to refresh reloads both lists', (tester) async {
    final server = FakeProfileServer()
      ..recommended = [publicUserJson(id: 'u-3', username: 'suggested_one')];
    final app = await _pumpPeople(tester, server);

    await tester.drag(find.text('suggested_one'), const Offset(0, 4000));
    await tester.pumpAndSettle();

    expect(app.requestsTo(ApiEndpoints.following), hasLength(2));
    expect(app.requestsTo(ApiEndpoints.recommendedChannels), hasLength(2));
  });

  testWidgets('a row opens the channel', (tester) async {
    final server = FakeProfileServer()
      ..recommended = [publicUserJson(id: 'u-2', username: 'other')];
    await _pumpPeople(tester, server);

    await tester.tap(find.byKey(const ValueKey('person-u-2')));
    await tester.pumpAndSettle();

    expect(find.byType(ChannelScreen), findsOneWidget);
  });
}
