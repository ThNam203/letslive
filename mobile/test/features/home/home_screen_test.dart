import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/home/home_providers.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';

import '../auth/auth_test_app.dart';
import 'home_fixtures.dart';
import 'home_test_app.dart';

/// Finder for something inside the section with [key].
Finder inSection(String key, Finder matching) =>
    find.descendant(of: find.byKey(ValueKey(key)), matching: matching);

void main() {
  testWidgets('home shows live and popular sections', (tester) async {
    final server = FakeHomeServer()
      ..live = [livestreamJson(title: 'Morning stream')]
      ..vods = [vodJson(title: 'Best of week')]
      ..following = [publicUserJson(id: 'u-2', username: 'followed_one')]
      ..recommended = [
        publicUserJson(id: 'u-2', username: 'followed_one'),
        publicUserJson(id: 'u-3', username: 'suggested_one'),
      ];
    await pumpHome(tester, server);

    expect(find.byType(HomeScreen), findsOneWidget);
    expect(inSection('section-live', find.text('Morning stream')), findsOne);
    expect(inSection('section-live', find.text('streamer')), findsOne);
    expect(
      inSection(
        'section-live',
        find.textContaining('1.5K views · Started 5 minutes ago'),
      ),
      findsOne,
    );
    expect(find.text('Best of week'), findsOne);
    expect(find.text('1:02:05'), findsOne);
    // The channel lists live in the profile's People tab, not on home.
    expect(find.byKey(const ValueKey('section-following')), findsNothing);
    expect(find.byKey(const ValueKey('section-recommended')), findsNothing);
    expect(find.text('followed_one'), findsNothing);
  });

  testWidgets('empty feed shows empty state', (tester) async {
    await pumpHome(tester, FakeHomeServer());

    expect(find.text('No livestreams'), findsOneWidget);
    expect(find.text('No Videos Available'), findsOneWidget);
  });

  testWidgets('error shows retry that refetches', (tester) async {
    final server = FakeHomeServer()
      ..liveStatus = 503
      ..vods = [vodJson(title: 'Still here')];
    final app = await pumpHome(tester, server);

    // One failing section does not take the others down.
    expect(inSection('section-live', find.text('Try again')), findsOne);
    expect(find.text('Still here'), findsOne);

    server
      ..liveStatus = 200
      ..live = [livestreamJson(title: 'Back online')];
    await tester.tap(inSection('section-live', find.text('Try again')));
    await tester.pumpAndSettle();

    expect(find.text('Back online'), findsOneWidget);
    expect(find.text('Try again'), findsNothing);
    expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(2));
    expect(app.requestsTo(ApiEndpoints.popularVods), hasLength(1));
  });

  testWidgets('pull to refresh reloads', (tester) async {
    final server = FakeHomeServer()..live = [livestreamJson(title: 'Old')];
    final app = await pumpHome(tester, server);
    expect(find.text('Old'), findsOneWidget);

    server.live = [livestreamJson(title: 'New')];
    // The indicator triggers after a drag of a quarter of the (tall) view.
    await tester.drag(find.byType(ListView).first, const Offset(0, 4000));
    await tester.pumpAndSettle();

    expect(find.text('New'), findsOneWidget);
    for (final path in [
      ApiEndpoints.popularLivestreams,
      ApiEndpoints.popularVods,
    ]) {
      expect(app.requestsTo(path), hasLength(2), reason: path);
    }
    // Owners are looked up again too, once per id.
    expect(app.requestsTo('${ApiEndpoints.user}/u-1'), hasLength(2));
  });

  group('coming back from the background', () {
    /// Moves the app through the given lifecycle states in order.
    void go(WidgetTester tester, List<AppLifecycleState> states) {
      for (final state in states) {
        tester.binding.handleAppLifecycleStateChanged(state);
      }
    }

    const leave = [
      AppLifecycleState.inactive,
      AppLifecycleState.hidden,
      AppLifecycleState.paused,
    ];
    const back = [
      AppLifecycleState.hidden,
      AppLifecycleState.inactive,
      AppLifecycleState.resumed,
    ];

    testWidgets('after a long time away refetches the feed', (tester) async {
      final server = FakeHomeServer()..live = [livestreamJson(title: 'Before')];
      var now = DateTime(2026, 11, 4, 12);
      final app = await pumpHome(
        tester,
        server,
        overrides: [homeClockProvider.overrideWithValue(() => now)],
      );

      server.live = [livestreamJson(title: 'After')];
      go(tester, leave);
      now = now.add(homeRefreshAfterBackground);
      go(tester, back);
      await tester.pumpAndSettle();

      expect(find.text('After'), findsOneWidget);
      expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(2));
    });

    testWidgets('after a short time away does not refetch', (tester) async {
      final app = await pumpHome(tester, FakeHomeServer());

      go(tester, leave);
      go(tester, back);
      await tester.pumpAndSettle();

      expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(1));
    });

    testWidgets('an inactive blip (control center) does not refetch', (
      tester,
    ) async {
      final app = await pumpHome(tester, FakeHomeServer());

      go(tester, [AppLifecycleState.inactive, AppLifecycleState.resumed]);
      await tester.pumpAndSettle();

      expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(1));
    });
  });

  testWidgets('cards of the same owner share one lookup', (tester) async {
    final server = FakeHomeServer()
      ..live = [livestreamJson(id: 'a'), livestreamJson(id: 'b')]
      ..vods = [vodJson(id: 'c'), vodJson(id: 'd')];
    final app = await pumpHome(tester, server);

    expect(find.text('streamer'), findsNWidgets(4));
    expect(app.requestsTo('${ApiEndpoints.user}/u-1'), hasLength(1));
  });

  testWidgets('a failed owner lookup still shows the card', (tester) async {
    final server = FakeHomeServer()
      ..users = {}
      ..live = [livestreamJson(title: 'Orphan stream')];
    await pumpHome(tester, server);

    expect(find.text('Orphan stream'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('very long texts do not overflow a small phone at 2x text', (
    tester,
  ) async {
    final longLive = 'Tiếng Việt rất dài 🎉 ' * 30;
    final longVod = 'Video dài lắm 🔥 ' * 30;
    final longOwner = 'Người_dùng_🎉${'x' * 200}';
    final server = FakeHomeServer()
      ..users = {'u-1': publicUserJson(username: longOwner)}
      ..live = [livestreamJson(title: longLive)]
      ..vods = [vodJson(title: longVod)]
      ..following = [publicUserJson(id: 'u-2', username: 'y' * 200)];
    await pumpHome(tester, server, width: 320, textScale: 2);

    expect(tester.takeException(), isNull);
    expect(find.textContaining('Tiếng Việt rất dài'), findsOneWidget);
    expect(find.textContaining('Video dài lắm'), findsOneWidget);
    expect(find.text(longOwner), findsNWidgets(2));
  });

  group('popular videos pagination', () {
    List<Map<String, Object?>> vodsOf(int n) => [
      for (var i = 0; i < n; i++) vodJson(id: 'v$i', title: 'Vod $i'),
    ];

    /// A narrow, very tall phone: every card of two pages gets built.
    Future<TestApp> pump(WidgetTester tester, FakeHomeServer server) =>
        pumpHome(tester, server, width: 320, height: 40000);

    List<Object?> pagesRequested(TestApp app) => [
      for (final r in app.requestsTo(ApiEndpoints.popularVods))
        r.queryParameters['page'],
    ];

    testWidgets('loads the next page near the end and stops at a short one', (
      tester,
    ) async {
      final server = FakeHomeServer()..vods = vodsOf(25);
      final app = await pump(tester, server);

      expect(find.text('Vod 0'), findsOneWidget);
      expect(find.text('Vod 24'), findsOneWidget);
      // Page 1 had 5 < 20 videos: the end, no third request.
      expect(pagesRequested(app), [0, 1]);
    });

    testWidgets('a full last page costs one empty request, then stops', (
      tester,
    ) async {
      final server = FakeHomeServer()..vods = vodsOf(20);
      final app = await pump(tester, server);

      expect(find.text('Vod 19'), findsOneWidget);
      expect(pagesRequested(app), [0, 1]);
    });

    testWidgets('builds cards lazily and loads more only near the end', (
      tester,
    ) async {
      final server = FakeHomeServer()..vods = vodsOf(45);
      final app = await pumpHome(tester, server, height: 700);

      expect(pagesRequested(app), [0]);
      expect(find.byType(MediaCard).evaluate().length, lessThan(8));

      for (var i = 0; i < 40 && pagesRequested(app).length < 2; i++) {
        await tester.drag(find.byType(Scrollable).first, const Offset(0, -400));
        await tester.pumpAndSettle();
      }

      expect(pagesRequested(app), [0, 1]);
    });

    testWidgets('a failed page shows an inline retry that keeps the list', (
      tester,
    ) async {
      final server = FakeHomeServer()
        ..vods = vodsOf(25)
        ..failVodsPage = 1;
      final app = await pump(tester, server);

      expect(find.text('Vod 19'), findsOneWidget);
      expect(find.text('Vod 20'), findsNothing);
      expect(find.text('Try again'), findsOneWidget);
      // No retry loop of its own.
      expect(pagesRequested(app), [0, 1]);

      server.failVodsPage = null;
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();

      expect(find.text('Vod 24'), findsOneWidget);
      expect(find.text('Try again'), findsNothing);
      expect(pagesRequested(app), [0, 1, 1]);
    });

    testWidgets('pull to refresh starts over at page 0', (tester) async {
      final server = FakeHomeServer()..vods = vodsOf(25);
      final app = await pump(tester, server);
      expect(pagesRequested(app), [0, 1]);

      unawaited(
        tester
            .state<RefreshIndicatorState>(find.byType(RefreshIndicator))
            .show(),
      );
      await tester.pumpAndSettle();

      // Reloaded from page 0; the footer then pages in the rest again.
      expect(pagesRequested(app), [0, 1, 0, 1]);
      expect(find.text('Vod 24'), findsOneWidget);
    });
  });

  testWidgets("logging in as another account shows none of the previous "
      "account's data", (tester) async {
    final server = FakeHomeServer()
      ..live = [livestreamJson(title: 'Stream of A')];
    final app = await pumpHome(tester, server);
    expect(find.text('Stream of A'), findsOneWidget);

    // Widget tests run in fake time: start the call, pump it through, await.
    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);

    server.live = [livestreamJson(title: 'Stream of B')];
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;

    expect(find.text('Stream of B'), findsOneWidget);
    expect(find.text('Stream of A'), findsNothing);
    expect(app.requestsTo(ApiEndpoints.popularLivestreams), hasLength(2));
  });
}
