import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';

import '../watch/fake_player.dart';
import 'vod_test_app.dart';

const _length = Duration(seconds: 3725);

/// A playing video at [seconds] of [_length].
PlayerValue _at(int seconds) => PlayerValue(
  isPlaying: true,
  isBuffering: false,
  position: Duration(seconds: seconds),
  duration: _length,
);

/// The VOD player's controls, wake lock and error retry.
void main() {
  testWidgets('shows the position and the length', (tester) async {
    final players = FakePlayers();
    await pumpVod(tester, FakeVodServer(), players);

    players.last.emit(_at(30));
    await tester.pumpAndSettle();

    expect(find.text('0:30 / 1:02:05'), findsOneWidget);
    expect(find.byTooltip('Full screen'), findsOneWidget);
  });

  testWidgets('picture-in-picture is in the settings menu', (tester) async {
    final players = FakePlayers();
    await pumpVod(tester, FakeVodServer(), players);

    await tester.tap(find.byKey(const ValueKey('player-settings')));
    await tester.pumpAndSettle();
    expect(find.text('Auto'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('player-pip')));
    await tester.pumpAndSettle();

    expect(players.last.pipRequests, 1);
  });

  group('skip buttons', () {
    testWidgets('jump 10 seconds either way', (tester) async {
      final players = FakePlayers();
      await pumpVod(tester, FakeVodServer(), players);
      players.last.emit(_at(30));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Forward 10 seconds'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Rewind 10 seconds'));
      await tester.pumpAndSettle();

      // The fake moves to the seeked position, so the second jump starts
      // from 40.
      expect(players.last.seeks, [
        const Duration(seconds: 40),
        const Duration(seconds: 30),
      ]);
    });

    testWidgets('clamp at the start and at the end', (tester) async {
      final players = FakePlayers();
      await pumpVod(tester, FakeVodServer(), players);

      players.last.emit(_at(5));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Rewind 10 seconds'));
      await tester.pumpAndSettle();

      players.last.emit(_at(3720));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Forward 10 seconds'));
      await tester.pumpAndSettle();

      expect(players.last.seeks, [Duration.zero, _length]);
    });
  });

  testWidgets('dragging the slider seeks when the finger lifts', (
    tester,
  ) async {
    final players = FakePlayers();
    await pumpVod(tester, FakeVodServer(), players);
    players.last.emit(_at(0));
    await tester.pumpAndSettle();

    await tester.drag(find.byType(Slider), const Offset(300, 0));
    await tester.pumpAndSettle();

    expect(players.last.seeks, hasLength(1));
    expect(players.last.seeks.single, greaterThan(const Duration(minutes: 5)));
    expect(players.last.seeks.single, lessThanOrEqualTo(_length));
  });

  testWidgets('controls fit at 320px with text scale 2', (tester) async {
    final players = FakePlayers();
    await pumpVod(
      tester,
      FakeVodServer(),
      players,
      width: 320,
      height: 568,
      textScale: 2,
    );

    players.last.emit(_at(3600));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.byType(Slider), findsOneWidget);
  });

  group('wake lock', () {
    testWidgets('is held while playing and released on leaving', (
      tester,
    ) async {
      final wake = FakeScreenWake();
      final players = FakePlayers();
      await pumpVod(tester, FakeVodServer(), players, wake: wake);

      players.last.emit(_at(1));
      await tester.pumpAndSettle();
      expect(wake.calls, [true]);

      await tester.binding.handlePopRoute(); // System back: no app bar here.
      await tester.pumpAndSettle();
      expect(wake.calls, [true, false]);
      expect(wake.isAwake, isFalse);
    });

    testWidgets('is released on pause', (tester) async {
      final wake = FakeScreenWake();
      final players = FakePlayers();
      await pumpVod(tester, FakeVodServer(), players, wake: wake);
      players.last.emit(_at(1));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Pause'));
      await tester.pumpAndSettle();

      expect(wake.isAwake, isFalse);
    });
  });

  group('playback error', () {
    PlayerValue failedAt(int seconds) => PlayerValue(
      isBuffering: false,
      hasError: true,
      position: Duration(seconds: seconds),
      duration: _length,
    );

    testWidgets('retry starts a second player where the viewer was', (
      tester,
    ) async {
      final players = FakePlayers();
      await pumpVod(tester, FakeVodServer(), players);
      players.last.emit(failedAt(120));
      await tester.pumpAndSettle();
      expect(find.text('Couldn\'t play the stream.'), findsOneWidget);
      expect(find.byType(Slider), findsNothing);

      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();

      expect(players.created, hasLength(2));
      expect(players.created.first.disposed, isTrue);
      expect(players.last.isLive, isFalse);
      expect(players.last.seeks, [const Duration(seconds: 120)]);
      expect(find.text('Couldn\'t play the stream.'), findsNothing);
    });

    testWidgets('an error before any playback retries from the start', (
      tester,
    ) async {
      final players = FakePlayers();
      await pumpVod(tester, FakeVodServer(), players);
      players.last.emit(failedAt(0));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Try again'));
      await tester.pumpAndSettle();

      expect(players.created, hasLength(2));
      expect(players.last.seeks, isEmpty);
    });
  });
}
