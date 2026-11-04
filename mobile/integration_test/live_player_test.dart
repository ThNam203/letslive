import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:letslive_mobile/features/watch/data/better_live_player_controller.dart';
import 'package:letslive_mobile/features/watch/domain/live_player_controller.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';
import 'package:wakelock_plus/wakelock_plus.dart';

/// HLS master playlist of a live stream to play, e.g. a local ffmpeg ladder
/// (see mobile/README.md "Video player"). The app derives the URL from the
/// livestream id; this test drives the real player directly so it can use
/// any stream.
const _url = String.fromEnvironment('LIVE_PLAYER_TEST_URL');

const _timeout = Duration(seconds: 30);

/// Pumps until [done] holds for the player's value.
Future<void> waitForValue(
  WidgetTester tester,
  LivePlayerController player,
  bool Function(LivePlayerValue value) done,
  String what,
) async {
  final deadline = DateTime.now().add(_timeout);
  while (!done(player.value.value)) {
    if (DateTime.now().isAfter(deadline)) {
      throw TestFailure('Timed out waiting for $what');
    }
    await tester.pump(const Duration(milliseconds: 250));
  }
}

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('real player plays a live HLS ladder and switches quality', (
    tester,
  ) async {
    expect(_url, isNotEmpty, reason: 'pass --dart-define=LIVE_PLAYER_TEST_URL');
    final player = BetterLivePlayerController(
      url: _url,
      title: 'Integration stream',
      author: 'integration',
    );
    addTearDown(player.dispose);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: AspectRatio(
            aspectRatio: 16 / 9,
            child: Builder(builder: player.buildVideo),
          ),
        ),
      ),
    );

    await waitForValue(
      tester,
      player,
      (v) => v.isPlaying && !v.isBuffering,
      'playback to start',
    );
    await waitForValue(
      tester,
      player,
      (v) => v.qualities.isNotEmpty,
      'the variants',
    );
    expect(player.value.value.qualities, const [
      VideoQuality(1080),
      VideoQuality(720),
      VideoQuality(360),
    ]);

    await player.selectQuality(const VideoQuality(360));
    expect(player.value.value.quality, const VideoQuality(360));
    await waitForValue(
      tester,
      player,
      (v) => v.isPlaying && !v.hasError,
      'playback after the switch',
    );

    await player.selectQuality(VideoQuality.auto);
    expect(player.value.value.quality, VideoQuality.auto);
    expect(player.value.value.hasError, isFalse);
  });

  testWidgets('the app screen wake reaches the native wakelock', (
    tester,
  ) async {
    final container = ProviderContainer();
    addTearDown(container.dispose);
    final wake = container.read(screenWakeProvider);

    await wake(true);
    expect(await WakelockPlus.enabled, isTrue);
    await wake(false);
    expect(await WakelockPlus.enabled, isFalse);
  });
}
