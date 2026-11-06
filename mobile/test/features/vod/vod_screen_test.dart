import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/vod/presentation/comments_section.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_player_view.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_screen.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';

import '../home/home_fixtures.dart';
import '../watch/fake_player.dart';
import '../watch/watch_test_app.dart';
import 'vod_test_app.dart';

/// The video screen: loading, error, not-found, the player, the details and
/// the view count.
void main() {
  testWidgets('shows the video: player, title, uploader, views, description', (
    tester,
  ) async {
    final server = FakeVodServer()
      ..vodById = {
        'vod-1': vodDetailJson(title: 'Best of week', description: 'About it'),
      };
    final players = FakePlayers();
    await pumpVod(tester, server, players);

    expect(find.byType(VodScreen), findsOneWidget);
    expect(find.text('Best of week'), findsWidgets);
    expect(find.text('About it'), findsOneWidget);
    expect(find.text('streamer'), findsWidgets);
    expect(find.textContaining('12 views'), findsOneWidget);
    // The URL is the video's playbackUrl, exactly as web feeds its player.
    expect(players.last.url, playbackUrl);
    expect(players.last.isLive, isFalse);
    expect(players.last.title, 'Best of week');
  });

  testWidgets('shows a spinner while the video loads', (tester) async {
    final server = FakeVodServer()..holdVod = Completer<void>();
    await pumpVod(tester, server, FakePlayers(), settle: false);

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    server.holdVod!.complete();
    await tester.pumpAndSettle();
    expect(find.text('Vod title'), findsWidgets);
  });

  testWidgets('a failed load shows the error with a retry that reloads', (
    tester,
  ) async {
    final server = FakeVodServer()..vodStatus = 503;
    final players = FakePlayers();
    await pumpVod(tester, server, players);

    expect(find.text('Try again'), findsOneWidget);
    expect(players.created, isEmpty);

    server.vodStatus = 200;
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.text('Try again'), findsNothing);
    expect(find.text('Vod title'), findsWidgets);
    expect(players.created, hasLength(1));
  });

  testWidgets('deleted vod shows not-found with back', (tester) async {
    final server = FakeVodServer()..vodById = {};
    final players = FakePlayers();
    await pumpVod(tester, server, players, vodId: 'gone');

    expect(find.text('VOD not found.'), findsOneWidget);
    expect(players.created, isEmpty);

    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(VodScreen), findsNothing);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('a deep link to a missing vod goes home from Back', (
    tester,
  ) async {
    final server = FakeVodServer()..vodById = {};
    final app = await pumpWatchHome(tester, server, FakePlayers());

    // `go` replaces the stack, like opening the app from a link: no pop.
    app.container.read(appRouterProvider).go(AppRoutes.vod('u-1', 'gone'));
    await tester.pumpAndSettle();
    expect(find.text('VOD not found.'), findsOneWidget);

    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(VodScreen), findsNothing);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('a vod of another user opens like on web, no owner check', (
    tester,
  ) async {
    final server = FakeVodServer()
      ..users = {
        'u-1': publicUserJson(),
        'u-9': publicUserJson(id: 'u-9', username: 'route_user'),
      };
    await pumpVod(tester, server, FakePlayers(), userId: 'u-9');

    // Web shows the video and the route's user as the uploader.
    expect(find.text('Vod title'), findsWidgets);
    expect(find.text('route_user'), findsOneWidget);
  });

  testWidgets('a vod without a playback url cannot play but still opens', (
    tester,
  ) async {
    final server = FakeVodServer()
      ..vodById = {'vod-1': vodDetailJson(url: null)};
    final players = FakePlayers();
    await pumpVod(tester, server, players);

    expect(players.created, isEmpty);
    expect(find.text('Vod title'), findsWidgets);
  });

  testWidgets('long description collapses with a more/less toggle', (
    tester,
  ) async {
    final long = List.filled(40, 'Một đoạn mô tả rất dài 🎉').join(' ');
    final server = FakeVodServer()
      ..vodById = {'vod-1': vodDetailJson(description: long)};
    await pumpVod(tester, server, FakePlayers());

    expect(find.text('...more'), findsOneWidget);
    await tester.tap(find.text('...more'));
    await tester.pumpAndSettle();
    expect(find.text('Show less'), findsOneWidget);
    await tester.tap(find.text('Show less'));
    await tester.pumpAndSettle();
    expect(find.text('...more'), findsOneWidget);
  });

  testWidgets('tapping a timestamp in the description seeks the player', (
    tester,
  ) async {
    final server = FakeVodServer()
      ..vodById = {'vod-1': vodDetailJson(description: 'Intro at 01:23 ok')};
    final players = FakePlayers();
    await pumpVod(tester, server, players);

    await tester.tapOnText(find.textRange.ofSubstring('01:23'));
    await tester.pumpAndSettle();

    expect(players.last.seeks, [const Duration(seconds: 83)]);
  });

  group('view count', () {
    PlayerValue at(int seconds) => PlayerValue(
      isPlaying: true,
      isBuffering: false,
      position: Duration(seconds: seconds),
      duration: const Duration(seconds: 3725),
    );

    testWidgets('is posted once, after 15 seconds of a long video', (
      tester,
    ) async {
      final server = FakeVodServer();
      final players = FakePlayers();
      await pumpVod(tester, server, players);

      players.last.emit(at(14));
      await tester.pumpAndSettle();
      expect(server.viewBodies, isEmpty);

      players.last.emit(at(15));
      await tester.pumpAndSettle();
      players.last.emit(at(16));
      await tester.pumpAndSettle();
      players.last.emit(at(900));
      await tester.pumpAndSettle();

      expect(server.viewBodies, [
        {'watchedSeconds': 15},
      ]);
    });

    testWidgets('needs only 10 percent of a short video', (tester) async {
      final server = FakeVodServer()
        ..vodById = {'vod-1': vodDetailJson(duration: 60)};
      final players = FakePlayers();
      await pumpVod(tester, server, players);

      players.last.emit(at(5));
      await tester.pumpAndSettle();
      expect(server.viewBodies, isEmpty);
      players.last.emit(at(6));
      await tester.pumpAndSettle();

      expect(server.viewBodies, [
        {'watchedSeconds': 6},
      ]);
    });

    testWidgets('a refused view is not retried on every tick', (tester) async {
      final server = FakeVodServer()..viewStatus = 400;
      final players = FakePlayers();
      await pumpVod(tester, server, players);

      for (var s = 20; s < 25; s++) {
        players.last.emit(at(s));
        await tester.pumpAndSettle();
      }
      expect(server.viewBodies, hasLength(1));
    });
  });

  testWidgets('leaving the screen disposes the player', (tester) async {
    final players = FakePlayers();
    await pumpVod(tester, FakeVodServer(), players);
    expect(players.last.disposed, isFalse);

    await tester.binding.handlePopRoute(); // System back: no app bar here.
    await tester.pumpAndSettle();
    expect(players.last.disposed, isTrue);
  });

  testWidgets('other streams lists the uploader\'s other videos', (
    tester,
  ) async {
    final server = FakeVodServer()
      ..userVods = [
        vodDetailJson(),
        vodDetailJson(id: 'vod-2', title: 'Another one'),
      ];
    await pumpVod(tester, server, FakePlayers());

    expect(find.text('Other streams'), findsOneWidget);
    expect(find.text('Another one'), findsOneWidget);

    await tester.tap(find.text('Another one'));
    await tester.pumpAndSettle();
    expect(find.byType(VodScreen), findsOneWidget);
  });

  testWidgets('long title and description fit at 320px, text scale 2', (
    tester,
  ) async {
    final long = 'Tiếng Việt không dấu và emoji 🎉' * 12;
    final server = FakeVodServer()
      ..vodById = {
        'vod-1': vodDetailJson(title: long, description: '$long 01:23 $long'),
      };
    await pumpVod(
      tester,
      server,
      FakePlayers(),
      width: 320,
      height: 568,
      textScale: 2,
    );

    expect(tester.takeException(), isNull);
    expect(find.byType(VodScreen), findsOneWidget);
  });

  testWidgets('landscape shows only the player, fitted; rotating back keeps '
      'it', (tester) async {
    final server = FakeVodServer()..vodById = {'vod-1': vodDetailJson()};
    final players = FakePlayers();
    await pumpVod(tester, server, players, width: 844, height: 390);

    expect(tester.takeException(), isNull);
    final player = tester.getRect(find.byType(VodPlayerView));
    expect(player.top, greaterThanOrEqualTo(0));
    expect(player.bottom, lessThanOrEqualTo(390));
    expect(player.height, closeTo(390, 1));
    expect(find.byType(AppBar), findsNothing);
    expect(find.byType(CommentsSection), findsNothing);

    tester.view.physicalSize = const Size(390, 844);
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    // No app bar upright either, like YouTube.
    expect(find.byType(AppBar), findsNothing);
    expect(find.byType(CommentsSection), findsOneWidget);
    // The same player: playback goes on across the rotation.
    expect(players.created, hasLength(1));
    expect(players.last.disposed, isFalse);
  });
}
