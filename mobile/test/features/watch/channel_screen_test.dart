import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';
import 'package:letslive_mobile/features/watch/domain/live_player_controller.dart';
import 'package:letslive_mobile/features/watch/presentation/channel_screen.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import '../home/home_fixtures.dart';
import 'fake_live_player.dart';
import 'watch_test_app.dart';

/// A server where `u-1` ("streamer") is live with `ls-1`.
FakeWatchServer liveServer({String title = 'Live title'}) =>
    FakeWatchServer()
      ..liveByUser = {'u-1': livestreamJson(id: 'ls-1', title: title)};

/// The quality menu entry labelled [label].
Finder menuItem(String label) =>
    find.widgetWithText(CheckedPopupMenuItem<VideoQuality>, label);

void main() {
  testWidgets('live user shows player with title', (tester) async {
    final players = FakeLivePlayers();
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
    await pumpChannel(tester, server, FakeLivePlayers(), 'u-1', settle: false);

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    server.holdLiveOf!.complete();
    await tester.pumpAndSettle();
    expect(find.text('Live title'), findsOneWidget);
  });

  testWidgets('offline user shows offline state with back action', (
    tester,
  ) async {
    final players = FakeLivePlayers();
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
      final players = FakeLivePlayers();
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
    final players = FakeLivePlayers();
    await pumpChannel(tester, server, players, 'u-1');

    server.liveByUser = {};
    players.last.emit(
      const LivePlayerValue(isBuffering: false, hasEnded: true),
    );
    await tester.pumpAndSettle();

    expect(find.text('This stream has ended.'), findsOneWidget);
  });

  testWidgets('playback error on a live stream offers a retry', (tester) async {
    final players = FakeLivePlayers();
    await pumpChannel(tester, liveServer(), players, 'u-1');

    players.last.emit(
      const LivePlayerValue(isBuffering: false, hasError: true),
    );
    await tester.pumpAndSettle();

    // Still live on the server, so it is a playback problem, not the end.
    expect(find.text("Couldn't play the stream."), findsOneWidget);
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(players.created, hasLength(2));
    expect(players.first.disposed, isTrue);
    expect(find.text("Couldn't play the stream."), findsNothing);
  });

  testWidgets('offline channel shows the player once the user goes live', (
    tester,
  ) async {
    final server = FakeWatchServer();
    final players = FakeLivePlayers();
    await pumpChannel(tester, server, players, 'u-1');
    expect(find.text('Offline'), findsOneWidget);

    server.liveByUser = {'u-1': livestreamJson(title: 'Now live')};
    await tester.pump(liveRecheckInterval);
    await tester.pumpAndSettle();

    expect(find.text('Now live'), findsOneWidget);
    expect(players.created, hasLength(1));
  });

  testWidgets('leaving the channel stops checking the stream', (tester) async {
    final players = FakeLivePlayers();
    final app = await pumpChannel(tester, liveServer(), players, 'u-1');

    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.pump(liveRecheckInterval * 3);
    await tester.pumpAndSettle();

    expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(1));
    expect(players.last.disposed, isTrue);
  });

  testWidgets('unknown user shows not-found with back action', (tester) async {
    final players = FakeLivePlayers();
    await pumpChannel(tester, FakeWatchServer(), players, 'missing');

    expect(find.text('User not found.'), findsOneWidget);
    expect(players.created, isEmpty);

    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('load error shows retry that refetches', (tester) async {
    final server = liveServer()..liveOfStatus = 503;
    final app = await pumpChannel(tester, server, FakeLivePlayers(), 'u-1');

    expect(find.text('Try again'), findsOneWidget);
    expect(find.text('Live title'), findsNothing);

    server.liveOfStatus = 200;
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.text('Live title'), findsOneWidget);
    expect(app.requestsTo(ApiEndpoints.livestreams), hasLength(2));
  });

  testWidgets('quality menu lists the stream variants', (tester) async {
    final players = FakeLivePlayers(
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
    final players = FakeLivePlayers();
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
      FakeLivePlayers(qualities: const [VideoQuality(1080), VideoQuality(720)]),
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
      FakeLivePlayers(),
      'u-1',
      width: 320,
      height: 700,
      textScale: 2,
    );
    expect(tester.takeException(), isNull);
    expect(find.text('Offline'), findsOneWidget);
  });
}

extension on FakeLivePlayers {
  FakeLivePlayer get first => created.first;
}
