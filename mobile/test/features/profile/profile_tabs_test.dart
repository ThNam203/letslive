import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_screen.dart';

import '../home/home_fixtures.dart';
import '../vod/vod_test_app.dart';
import 'profile_test_app.dart';

Future<void> _openTab(WidgetTester tester, String label) async {
  await tester.tap(find.widgetWithText(Tab, label));
  await tester.pumpAndSettle();
}

/// The profile of a channel: header and the Home / Videos / About tabs.
void main() {
  testWidgets('header shows the user and the Chat tab is the default', (
    tester,
  ) async {
    final server = FakeProfileServer();
    await pumpRoute(tester, server, AppRoutes.channel('u-2'));

    expect(find.text('other'), findsWidgets);
    expect(find.text('7 followers'), findsOneWidget);
    expect(find.text('Joined Jan 15, 2026'), findsOneWidget);
    for (final tab in ['Chat', 'Home', 'Videos', 'About']) {
      expect(find.widgetWithText(Tab, tab), findsOneWidget, reason: tab);
    }
    expect(find.text('Type a message...'), findsOneWidget);
  });

  group('Home tab', () {
    testWidgets('features the latest video and lists more', (tester) async {
      final server = FakeProfileServer()..channelVods = vodsOf(7);
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await _openTab(tester, 'Home');

      expect(find.text('Latest video'), findsOneWidget);
      expect(
        find.descendant(
          of: find.byKey(const ValueKey('featured-vod')),
          matching: find.text('Video 0'),
        ),
        findsOneWidget,
      );
      expect(find.text('More videos'), findsOneWidget);
      // Four more at most, like web.
      expect(find.byKey(const ValueKey('vod-v-4')), findsOneWidget);
      expect(find.byKey(const ValueKey('vod-v-5')), findsNothing);
    });

    testWidgets('View all opens the Videos tab, Watch opens the video', (
      tester,
    ) async {
      final server = FakeProfileServer()
        ..channelVods = vodsOf(3)
        ..vodById = {'v-0': vodDetailJson(id: 'v-0', userId: 'u-2')};
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await _openTab(tester, 'Home');

      await tester.tap(find.byKey(const ValueKey('view-all-videos')));
      await tester.pumpAndSettle();
      expect(find.text('Latest video'), findsNothing);
      expect(find.byKey(const ValueKey('vod-v-0')), findsOneWidget);

      await _openTab(tester, 'Home');
      await tester.tap(find.text('Watch'));
      await tester.pumpAndSettle();
      expect(find.byType(VodScreen), findsOneWidget);
    });

    testWidgets('no videos says so', (tester) async {
      await pumpRoute(tester, FakeProfileServer(), AppRoutes.channel('u-2'));
      await _openTab(tester, 'Home');

      expect(find.text('No videos yet'), findsOneWidget);
      expect(find.textContaining("other hasn't published"), findsOneWidget);
    });

    testWidgets('a failed load shows retry that loads again', (tester) async {
      final server = FakeProfileServer()
        ..channelVods = vodsOf(2)
        ..channelVodsStatus = 503;
      final app = await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await _openTab(tester, 'Home');
      expect(find.text('Try again'), findsOneWidget);

      server.channelVodsStatus = 200;
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();
      expect(find.text('Latest video'), findsOneWidget);
      expect(app.requestsTo(ApiEndpoints.vods), hasLength(2));
    });
  });

  group('Videos tab', () {
    testWidgets('loads the next page when scrolled to the end', (tester) async {
      final server = FakeProfileServer()
        ..channelVods = vodsOf(channelVodsPageSize + 3);
      final app = await pumpRoute(
        tester,
        server,
        AppRoutes.channel('u-2'),
        height: 700,
      );
      await _openTab(tester, 'Videos');
      expect(app.requestsTo(ApiEndpoints.vods), hasLength(1));
      expect(
        app.requestsTo(ApiEndpoints.vods).single.queryParameters['limit'],
        channelVodsPageSize,
      );

      final last = find.byKey(ValueKey('vod-v-${channelVodsPageSize + 2}'));
      await tester.dragUntilVisible(
        last,
        find.byType(ListView),
        const Offset(0, -300),
        maxIteration: 80,
      );
      await tester.pumpAndSettle();

      expect(app.requestsTo(ApiEndpoints.vods), hasLength(2));
      expect(app.requestsTo(ApiEndpoints.vods).last.queryParameters['page'], 1);
      expect(last, findsOneWidget);
    });

    testWidgets('a failed next page keeps the list and retries inline', (
      tester,
    ) async {
      final server = FakeProfileServer()
        ..channelVods = vodsOf(channelVodsPageSize + 1)
        ..failChannelVodsPage = 1;
      await pumpRoute(tester, server, AppRoutes.channel('u-2'), height: 700);
      await _openTab(tester, 'Videos');

      await tester.dragUntilVisible(
        find.text('Try again'),
        find.byType(ListView),
        const Offset(0, -300),
        maxIteration: 80,
      );

      server.failChannelVodsPage = null;
      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('vod-v-$channelVodsPageSize')),
        findsOneWidget,
      );
    });

    testWidgets('a card opens the video', (tester) async {
      final server = FakeProfileServer()
        ..channelVods = vodsOf(2)
        ..vodById = {'v-1': vodDetailJson(id: 'v-1', userId: 'u-2')};
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await _openTab(tester, 'Videos');

      await tester.tap(find.text('Video 1'));
      await tester.pumpAndSettle();
      expect(find.byType(VodScreen), findsOneWidget);
    });
  });

  group('About tab', () {
    testWidgets('shows bio, stats and links; a tap opens the browser', (
      tester,
    ) async {
      final server = FakeProfileServer();
      server.users['u-2'] = publicUserJson(
        id: 'u-2',
        username: 'other',
        bio: 'Hello there',
        socialMediaLinks: {
          'github': 'https://www.github.com/octocat',
          'website': 'javascript:alert(1)',
          'youtube': 'ftp://example.com/x',
        },
      );
      final device = FakeDevice();
      await pumpRoute(tester, server, AppRoutes.channel('u-2'), device: device);
      await _openTab(tester, 'About');

      expect(find.text('About other'), findsOneWidget);
      expect(find.text('Hello there'), findsOneWidget);
      expect(find.text('Channel info'), findsOneWidget);
      expect(find.text('github.com/octocat'), findsOneWidget);
      // Only http(s) links are listed.
      expect(find.textContaining('alert'), findsNothing);
      expect(find.textContaining('example.com'), findsNothing);

      await tester.tap(find.text('github.com/octocat'));
      await tester.pump();
      expect(device.opened, [Uri.parse('https://www.github.com/octocat')]);
    });

    testWidgets('a link nothing can open tells the user', (tester) async {
      final server = FakeProfileServer();
      server.users['u-2'] = publicUserJson(
        id: 'u-2',
        username: 'other',
        socialMediaLinks: {'github': 'https://github.com/octocat'},
      );
      final device = FakeDevice()..canOpen = false;
      await pumpRoute(tester, server, AppRoutes.channel('u-2'), device: device);
      await _openTab(tester, 'About');

      await tester.tap(find.text('github.com/octocat'));
      await tester.pump();
      expect(
        find.text('Something went wrong. Please try again.'),
        findsOneWidget,
      );
    });

    testWidgets('no bio shows the placeholder', (tester) async {
      await pumpRoute(tester, FakeProfileServer(), AppRoutes.channel('u-2'));
      await _openTab(tester, 'About');

      expect(find.text("other hasn't written a bio yet."), findsOneWidget);
    });
  });

  testWidgets('long names and bios fit 320px at text scale 2 in every tab', (
    tester,
  ) async {
    final longName = 'tên_rất_dài_${'x' * 60}';
    final server = FakeProfileServer()..channelVods = vodsOf(6);
    server.users['u-2'] = publicUserJson(
      id: 'u-2',
      username: longName,
      bio: 'Xin chào 🎉 ${'tiểu sử rất dài không dấu ' * 20}',
      socialMediaLinks: {'website': 'https://example.com/${'p' * 80}'},
    );
    await pumpRoute(
      tester,
      server,
      AppRoutes.channel('u-2'),
      width: 320,
      height: 700,
      textScale: 2,
    );
    expect(tester.takeException(), isNull, reason: 'Chat');

    for (final tab in ['Home', 'Videos', 'About']) {
      await _openTab(tester, tab);
      expect(tester.takeException(), isNull, reason: tab);
    }
  });

  testWidgets('the unfollow button fits 320px at text scale 2 in Vietnamese', (
    tester,
  ) async {
    final server = FakeProfileServer();
    server.users['u-2'] = publicUserJson(
      id: 'u-2',
      username: 'tên_rất_dài_${'x' * 60}',
      followerCount: 1234567,
      isFollowing: true,
    );
    final device = FakeDevice()..locale = const Locale('vi');
    await pumpRoute(
      tester,
      server,
      AppRoutes.channel('u-2'),
      device: device,
      width: 320,
      height: 700,
      textScale: 2,
    );

    expect(find.text('Bỏ theo dõi'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('leaving the Chat tab keeps its draft and history', (
    tester,
  ) async {
    final server = FakeProfileServer();
    final app = await pumpRoute(tester, server, AppRoutes.channel('u-2'));
    final input = find.byType(TextField, skipOffstage: false);
    await tester.enterText(input, 'half written');

    await _openTab(tester, 'Home');
    await _openTab(tester, 'About');
    await tester.tap(find.widgetWithText(Tab, 'Chat'));
    await tester.pumpAndSettle();

    expect(tester.widget<TextField>(input).controller!.text, 'half written');
    final history = app
        .requestsTo(ApiEndpoints.messages)
        .where((r) => r.method == 'GET');
    expect(history, hasLength(1));
  });
}
