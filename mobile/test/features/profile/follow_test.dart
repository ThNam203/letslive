import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';

import '../home/home_fixtures.dart';
import '../vod/vod_test_app.dart';
import 'profile_test_app.dart';

final _follow = find.byKey(const ValueKey('follow-button'));

/// Follow and unfollow on a channel and on the video's uploader row.
void main() {
  testWidgets('follow updates count immediately and rolls back on error', (
    tester,
  ) async {
    final server = FakeProfileServer()
      ..holdFollow = Completer<void>()
      ..followStatus = 503;
    await pumpRoute(tester, server, AppRoutes.channel('u-2'));
    expect(find.text('7 followers'), findsOneWidget);
    expect(find.text('Follow'), findsOneWidget);

    await tester.tap(_follow);
    await tester.pump();
    // Shown at once, before the server answered.
    expect(find.text('8 followers'), findsOneWidget);
    expect(find.text('Unfollow'), findsOneWidget);

    server.holdFollow!.complete();
    await tester.pumpAndSettle();
    expect(find.text('7 followers'), findsOneWidget);
    expect(find.text('Follow'), findsOneWidget);
    expect(find.text('Something went wrong.'), findsOneWidget);
  });

  testWidgets('a second tap while the request is in flight is ignored', (
    tester,
  ) async {
    final server = FakeProfileServer()..holdFollow = Completer<void>();
    await pumpRoute(tester, server, AppRoutes.channel('u-2'));

    await tester.tap(_follow);
    await tester.pump();
    await tester.tap(_follow);
    await tester.pump();
    server.holdFollow!.complete();
    await tester.pumpAndSettle();

    expect(server.followCalls, ['follow:u-2']);
    expect(find.text('8 followers'), findsOneWidget);
  });

  testWidgets('unfollowing sends DELETE and lowers the count', (tester) async {
    final server = FakeProfileServer();
    server.users['u-2'] = publicUserJson(
      id: 'u-2',
      username: 'other',
      isFollowing: true,
    );
    await pumpRoute(tester, server, AppRoutes.channel('u-2'));
    expect(find.text('Unfollow'), findsOneWidget);

    await tester.tap(_follow);
    await tester.pumpAndSettle();

    expect(server.followCalls, ['unfollow:u-2']);
    expect(find.text('6 followers'), findsOneWidget);
    expect(find.text('Follow'), findsOneWidget);
  });

  testWidgets('your own channel offers no follow button', (tester) async {
    await pumpRoute(tester, FakeProfileServer(), AppRoutes.channel('u-1'));

    expect(find.text('abcdef'), findsWidgets);
    expect(_follow, findsNothing);
  });

  testWidgets('following refreshes the People tab\'s following list', (
    tester,
  ) async {
    final server = FakeProfileServer();
    final app = await pumpRoute(tester, server, AppRoutes.profile);
    await tester.tap(find.widgetWithText(Tab, 'People'));
    await tester.pumpAndSettle();
    expect(app.requestsTo(ApiEndpoints.following), hasLength(1));

    unawaited(
      app.container.read(appRouterProvider).push(AppRoutes.channel('u-2')),
    );
    await tester.pumpAndSettle();
    await tester.tap(_follow);
    await tester.pumpAndSettle();

    // The tab is paused under the channel; it reloads when it shows again.
    await tester.binding.handlePopRoute(); // System back: no app bar here.
    await tester.pumpAndSettle();
    expect(app.requestsTo(ApiEndpoints.following), hasLength(2));
  });

  testWidgets('the video screen has a follow button on the uploader row', (
    tester,
  ) async {
    final server = FakeProfileServer()
      ..vodById = {'vod-2': vodDetailJson(id: 'vod-2', userId: 'u-2')};
    await pumpRoute(tester, server, AppRoutes.vod('u-2', 'vod-2'));

    await tester.tap(_follow);
    await tester.pumpAndSettle();

    expect(server.followCalls, ['follow:u-2']);
    expect(find.text('Unfollow'), findsOneWidget);
  });
}
