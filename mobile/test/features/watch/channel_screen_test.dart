import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/chat_panel.dart';
import 'package:letslive_mobile/features/watch/presentation/channel_screen.dart';
import 'package:letslive_mobile/features/watch/presentation/live_player_view.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import '../home/home_fixtures.dart';
import 'fake_player.dart';
import 'watch_test_app.dart';

/// A server where `u-1` ("streamer") is live with `ls-1`.
FakeWatchServer liveServer({String title = 'Live title'}) =>
    FakeWatchServer()
      ..liveByUser = {'u-1': livestreamJson(id: 'ls-1', title: title)};

/// The native player reached the end of the playlist.
const ended = PlayerValue(isBuffering: false, hasEnded: true);

/// The native player failed.
const failed = PlayerValue(isBuffering: false, hasError: true);

/// The quality menu entry labelled [label].
Finder menuItem(String label) =>
    find.widgetWithText(CheckedPopupMenuItem<VideoQuality>, label);

void main() {
  testWidgets('live user shows player with title', (tester) async {
    final players = FakePlayers();
    await pumpChannel(tester, liveServer(), players, 'u-1');

    expect(find.byType(ChannelScreen), findsOneWidget);
    expect(find.byKey(const ValueKey('fake-video')), findsOneWidget);
    expect(find.text('Live title'), findsOneWidget);
    expect(find.text('streamer'), findsWidgets);
    expect(find.text('Live'), findsOneWidget);
    // The HLS URL is built exactly like web's channel page.
    expect(players.last.url, 'http://localhost:8000/transcode/ls-1/index.m3u8');
    // The background-playback notification names the stream and streamer.
    expect(players.last.title, 'Live title');
    expect(players.last.author, 'streamer');
  });

  testWidgets('shows a spinner while the channel loads', (tester) async {
    final server = liveServer()..holdLiveOf = Completer<void>();
    await pumpChannel(tester, server, FakePlayers(), 'u-1', settle: false);

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    server.holdLiveOf!.complete();
    await tester.pumpAndSettle();
    expect(find.text('Live title'), findsOneWidget);
  });

  testWidgets('offline user shows offline state with back action', (
    tester,
  ) async {
    final players = FakePlayers();
    await pumpChannel(tester, FakeWatchServer(), players, 'u-1');

    expect(find.text('Offline'), findsOneWidget);
    expect(find.text("streamer isn't live right now."), findsOneWidget);
    expect(players.created, isEmpty);

    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(ChannelScreen), findsNothing);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets(
    'stream ended while watching shows ended state with back action',
    (tester) async {
      final server = liveServer();
      final players = FakePlayers();
      final app = await pumpChannel(tester, server, players, 'u-1');
      expect(find.byKey(const ValueKey('fake-video')), findsOneWidget);

      // The streamer stops; the next check finds no live stream.
      server.liveByUser = {};
      await tester.pump(liveRecheckInterval);
      await tester.pumpAndSettle();

      expect(find.text('This stream has ended.'), findsOneWidget);
      expect(find.byKey(const ValueKey('fake-video')), findsNothing);
      expect(players.last.disposed, isTrue);
      expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(2));

      await tester.tap(find.text('Back'));
      await tester.pumpAndSettle();
      expect(find.byType(HomeScreen), findsOneWidget);
    },
  );

  testWidgets('player reaching the end re-checks the stream at once', (
    tester,
  ) async {
    final server = liveServer();
    final players = FakePlayers();
    await pumpChannel(tester, server, players, 'u-1');

    server.liveByUser = {};
    players.last.emit(ended);
    await tester.pumpAndSettle();

    expect(find.text('This stream has ended.'), findsOneWidget);
  });

  testWidgets('ended playback shows a spinner, not an error, until checked', (
    tester,
  ) async {
    final server = liveServer();
    final players = FakePlayers();
    await pumpChannel(tester, server, players, 'u-1');

    server.holdLiveOf = Completer<void>();
    server.liveByUser = {};
    players.last.emit(ended);
    await tester.pump();

    expect(find.text("Couldn't play the stream."), findsNothing);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    server.holdLiveOf!.complete();
    await tester.pumpAndSettle();
    expect(find.text('This stream has ended.'), findsOneWidget);
  });

  testWidgets(
    'ended playback on a still-live stream waits for the next check',
    (tester) async {
      final server = liveServer();
      final players = FakePlayers();
      await pumpChannel(tester, server, players, 'u-1');

      // The playlist ended but the server still lists the stream.
      players.last.emit(ended);
      await tester.pump();
      await tester.pump();
      expect(find.text("Couldn't play the stream."), findsNothing);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      server.liveByUser = {};
      await tester.pump(liveRecheckInterval);
      await tester.pumpAndSettle();
      expect(find.text('This stream has ended.'), findsOneWidget);
    },
  );

  testWidgets('first playback error restarts the player without a message', (
    tester,
  ) async {
    final players = FakePlayers();
    final app = await pumpChannel(tester, liveServer(), players, 'u-1');

    players.last.emit(failed);
    await tester.pumpAndSettle();

    expect(players.created, hasLength(2));
    expect(players.first.disposed, isTrue);
    expect(find.text("Couldn't play the stream."), findsNothing);
    expect(find.byKey(const ValueKey('fake-video')), findsOneWidget);
    // A silent restart does not ask the server.
    expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(1));
  });

  testWidgets('second playback error re-checks, then offers a retry', (
    tester,
  ) async {
    final server = liveServer();
    final players = FakePlayers();
    final app = await pumpChannel(tester, server, players, 'u-1');
    players.last.emit(failed);
    await tester.pumpAndSettle();

    // The restarted player fails before it ever plays.
    server.holdLiveOf = Completer<void>();
    players.last.emit(failed);
    await tester.pump();
    expect(find.text("Couldn't play the stream."), findsNothing);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    // Still live on the server, so it is a playback problem, not the end.
    server.holdLiveOf!.complete();
    await tester.pumpAndSettle();
    expect(find.text("Couldn't play the stream."), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(2));

    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();
    expect(players.created, hasLength(3));
    expect(players.created[1].disposed, isTrue);
    expect(find.text("Couldn't play the stream."), findsNothing);
  });

  testWidgets('playing again re-arms the automatic restart', (tester) async {
    final players = FakePlayers();
    await pumpChannel(tester, liveServer(), players, 'u-1');
    players.last.emit(failed);
    await tester.pumpAndSettle();

    players.last.emit(const PlayerValue(isPlaying: true, isBuffering: false));
    players.last.emit(failed);
    await tester.pumpAndSettle();

    expect(players.created, hasLength(3));
    expect(find.text("Couldn't play the stream."), findsNothing);
  });

  testWidgets('offline channel shows the player once the user goes live', (
    tester,
  ) async {
    final server = FakeWatchServer();
    final players = FakePlayers();
    await pumpChannel(tester, server, players, 'u-1');
    expect(find.text('Offline'), findsOneWidget);

    server.liveByUser = {'u-1': livestreamJson(title: 'Now live')};
    await tester.pump(liveRecheckInterval);
    await tester.pumpAndSettle();

    expect(find.text('Now live'), findsOneWidget);
    expect(players.created, hasLength(1));
  });

  testWidgets('leaving the channel stops checking the stream', (tester) async {
    final players = FakePlayers();
    final app = await pumpChannel(tester, liveServer(), players, 'u-1');

    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.pump(liveRecheckInterval * 3);
    await tester.pumpAndSettle();

    expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(1));
    expect(players.last.disposed, isTrue);
  });

  testWidgets('unknown user shows not-found with back action', (tester) async {
    final players = FakePlayers();
    await pumpChannel(tester, FakeWatchServer(), players, 'missing');

    expect(find.text('User not found.'), findsOneWidget);
    expect(players.created, isEmpty);

    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('load error shows retry that refetches', (tester) async {
    final server = liveServer()..liveOfStatus = 503;
    final app = await pumpChannel(tester, server, FakePlayers(), 'u-1');

    expect(find.text('Try again'), findsOneWidget);
    expect(find.text('Live title'), findsNothing);

    server.liveOfStatus = 200;
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.text('Live title'), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(2));
  });

  testWidgets('quality menu lists the stream variants', (tester) async {
    final players = FakePlayers(
      qualities: const [
        VideoQuality(1080),
        VideoQuality(720),
        VideoQuality(360),
      ],
    );
    await pumpChannel(tester, liveServer(), players, 'u-1');

    await tester.tap(find.byKey(const ValueKey('player-quality')));
    await tester.pumpAndSettle();
    for (final label in ['Auto', '1080p', '720p', '360p']) {
      expect(find.text(label), findsOneWidget, reason: label);
    }

    await tester.tap(menuItem('720p'));
    await tester.pumpAndSettle();
    expect(players.last.selected, [const VideoQuality(720)]);

    // Back to adaptive.
    await tester.tap(find.byKey(const ValueKey('player-quality')));
    await tester.pumpAndSettle();
    await tester.tap(menuItem('Auto'));
    await tester.pumpAndSettle();
    expect(players.last.selected.last, VideoQuality.auto);
  });

  testWidgets('play/pause and picture-in-picture reach the player', (
    tester,
  ) async {
    final players = FakePlayers();
    await pumpChannel(tester, liveServer(), players, 'u-1');

    await tester.tap(find.byTooltip('Pause'));
    await tester.pump();
    expect(players.last.value.value.isPlaying, isFalse);
    await tester.tap(find.byTooltip('Play'));
    await tester.pump();
    expect(players.last.value.value.isPlaying, isTrue);

    await tester.tap(find.byTooltip('Picture in picture'));
    await tester.pump();
    expect(players.last.pipRequests, 1);
  });

  testWidgets('long titles and names do not overflow on a small phone', (
    tester,
  ) async {
    final longName = 'người_xem_rất_dài_${'x' * 60}';
    final server = FakeWatchServer()
      ..users = {'u-1': publicUserJson(username: longName)}
      ..liveByUser = {
        'u-1': livestreamJson(
          title: 'Trò chơi 🎮🔥 ${'Phát trực tiếp không dấu ' * 12}',
        ),
      };
    await pumpChannel(
      tester,
      server,
      FakePlayers(qualities: const [VideoQuality(1080), VideoQuality(720)]),
      'u-1',
      width: 320,
      height: 700,
      textScale: 2,
    );

    expect(tester.takeException(), isNull);
    expect(find.byKey(const ValueKey('fake-video')), findsOneWidget);

    server.liveByUser = {};
    await tester.pump(liveRecheckInterval);
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.text('This stream has ended.'), findsOneWidget);
  });

  testWidgets('offline and not-found states fit a small phone', (tester) async {
    final server = FakeWatchServer()
      ..users = {'u-1': publicUserJson(username: 'tên_${'x' * 80}')};
    await pumpChannel(
      tester,
      server,
      FakePlayers(),
      'u-1',
      width: 320,
      height: 700,
      textScale: 2,
    );
    expect(tester.takeException(), isNull);
    expect(find.text('Offline'), findsOneWidget);
  });

  testWidgets('landscape shows only the live player, fitted; rotating back '
      'keeps it', (tester) async {
    final players = FakePlayers();
    await pumpChannel(
      tester,
      liveServer(),
      players,
      'u-1',
      width: 844,
      height: 390,
    );

    expect(tester.takeException(), isNull);
    final player = tester.getRect(find.byType(LivePlayerView));
    expect(player.top, greaterThanOrEqualTo(0));
    expect(player.bottom, lessThanOrEqualTo(390));
    expect(player.height, closeTo(390, 1));
    expect(find.byType(AppBar), findsNothing);
    expect(find.byType(ChatPanel), findsNothing);

    tester.view.physicalSize = const Size(390, 844);
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.byType(AppBar), findsOneWidget);
    expect(find.byType(ChatPanel), findsOneWidget);
    expect(players.created, hasLength(1));
    expect(players.last.disposed, isFalse);
  });

  group('screen wake', screenWakeTests);
}

void screenWakeTests() {
  testWidgets('playing keeps the screen on, pause lets it sleep', (
    tester,
  ) async {
    final wake = FakeScreenWake();
    await pumpChannel(tester, liveServer(), FakePlayers(), 'u-1', wake: wake);
    expect(wake.isAwake, isTrue);

    await tester.tap(find.byTooltip('Pause'));
    await tester.pump();
    expect(wake.isAwake, isFalse);

    await tester.tap(find.byTooltip('Play'));
    await tester.pump();
    expect(wake.isAwake, isTrue);
    expect(wake.calls, [true, false, true]);
  });

  testWidgets('a playback error or the end lets the screen sleep', (
    tester,
  ) async {
    final wake = FakeScreenWake();
    final players = FakePlayers();
    final server = liveServer();
    await pumpChannel(tester, server, players, 'u-1', wake: wake);

    players.last.emit(failed);
    expect(wake.isAwake, isFalse);
    // The silent restart plays again.
    await tester.pumpAndSettle();
    expect(wake.isAwake, isTrue);

    server.liveByUser = {};
    players.last.emit(ended);
    expect(wake.isAwake, isFalse);
    await tester.pumpAndSettle();
    expect(find.text('This stream has ended.'), findsOneWidget);
    expect(wake.isAwake, isFalse);
  });

  testWidgets('leaving the channel lets the screen sleep', (tester) async {
    final wake = FakeScreenWake();
    await pumpChannel(tester, liveServer(), FakePlayers(), 'u-1', wake: wake);
    expect(wake.isAwake, isTrue);

    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(wake.isAwake, isFalse);
  });

  testWidgets('logging out lets the screen sleep', (tester) async {
    final wake = FakeScreenWake();
    final app = await pumpChannel(
      tester,
      liveServer(),
      FakePlayers(),
      'u-1',
      wake: wake,
    );
    expect(wake.isAwake, isTrue);

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);
    expect(wake.isAwake, isFalse);
  });
}

extension on FakePlayers {
  FakePlayer get first => created.first;
}
