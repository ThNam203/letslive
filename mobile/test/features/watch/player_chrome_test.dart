import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:letslive_mobile/features/watch/presentation/live_player_view.dart';
import 'package:letslive_mobile/features/watch/presentation/player_chrome.dart';

import 'fake_player.dart';

/// A playing [FakePlayer] in a [PlayerChrome] whose bar has the play/pause
/// and full-screen buttons, on a [width] x [height] screen.
Future<FakePlayer> _pump(
  WidgetTester tester, {
  double width = 400,
  double height = 800,
}) async {
  tester.view.physicalSize = Size(width, height);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  final player = FakePlayer(url: 'u', title: 't', author: 'a');
  await tester.pumpWidget(
    MaterialApp(
      locale: const Locale('en'),
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ],
      home: Scaffold(
        body: AspectRatio(
          aspectRatio: 16 / 9,
          child: ValueListenableBuilder<PlayerValue>(
            valueListenable: player.value,
            builder: (context, value, _) => PlayerChrome(
              player: player,
              value: value,
              bar: Row(
                children: [
                  PlayPauseButton(player: player, value: value),
                  const FullscreenButton(),
                ],
              ),
            ),
          ),
        ),
      ),
    ),
  );
  return player;
}

/// Opacity of the control bar (1 shown, 0 faded out).
double _controls(WidgetTester tester) => tester
    .widget<AnimatedOpacity>(find.byKey(const ValueKey('player-controls')))
    .opacity;

/// Records the orientations the app asks the system for.
List<List<String>> _recordOrientations(WidgetTester tester) {
  final asked = <List<String>>[];
  tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
    SystemChannels.platform,
    (call) async {
      if (call.method == 'SystemChrome.setPreferredOrientations') {
        asked.add((call.arguments as List<Object?>).cast<String>());
      }
      return null;
    },
  );
  addTearDown(
    () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      SystemChannels.platform,
      null,
    ),
  );
  return asked;
}

void main() {
  testWidgets('tapping the video pauses it, tapping again plays it', (
    tester,
  ) async {
    final player = await _pump(tester);

    await tester.tap(find.byKey(const ValueKey('player-surface')));
    await tester.pump();
    expect(player.value.value.isPlaying, isFalse);

    await tester.tap(find.byKey(const ValueKey('player-surface')));
    await tester.pump();
    expect(player.value.value.isPlaying, isTrue);
  });

  testWidgets('the controls fade out 3 seconds after the last touch while '
      'playing, and a touch brings them back', (tester) async {
    await _pump(tester);
    expect(_controls(tester), 1);

    await tester.pump(controlsHideDelay - const Duration(milliseconds: 100));
    expect(_controls(tester), 1);
    await tester.pump(const Duration(milliseconds: 200));
    await tester.pumpAndSettle();
    expect(_controls(tester), 0);

    // A touch shows them (and, on the video, pauses: paused keeps them up).
    await tester.tap(find.byKey(const ValueKey('player-surface')));
    await tester.pump();
    expect(_controls(tester), 1);
    await tester.pump(controlsHideDelay * 2);
    expect(_controls(tester), 1);
  });

  testWidgets('a finger held on the controls keeps them up', (tester) async {
    await _pump(tester);
    final finger = await tester.startGesture(
      tester.getCenter(find.byKey(const ValueKey('player-fullscreen'))) +
          const Offset(60, 0),
    );
    await tester.pump(controlsHideDelay * 2);
    expect(_controls(tester), 1);

    await finger.up();
    await tester.pump(controlsHideDelay + const Duration(milliseconds: 100));
    await tester.pumpAndSettle();
    expect(_controls(tester), 0);
  });

  testWidgets('hidden controls let a tap through to the video', (tester) async {
    final player = await _pump(tester);
    await tester.pump(controlsHideDelay);
    await tester.pumpAndSettle();
    expect(_controls(tester), 0);

    // Where the pause button was: the tap reaches the video instead.
    await tester.tap(find.byTooltip('Pause'), warnIfMissed: false);
    await tester.pump();
    expect(player.value.value.isPlaying, isFalse);
    expect(_controls(tester), 1);
  });

  testWidgets('the full-screen button turns the screen sideways', (
    tester,
  ) async {
    final asked = _recordOrientations(tester);
    await _pump(tester);

    await tester.tap(find.byTooltip('Full screen'));
    await tester.pump();

    expect(asked, [
      ['DeviceOrientation.landscapeLeft', 'DeviceOrientation.landscapeRight'],
    ]);
  });

  testWidgets('sideways the button leaves full screen, upright', (
    tester,
  ) async {
    final asked = _recordOrientations(tester);
    await _pump(tester, width: 844, height: 390);
    expect(isLandscape(tester.element(find.byType(PlayerChrome))), isTrue);

    await tester.tap(find.byTooltip('Exit full screen'));
    await tester.pump();

    expect(asked, [
      ['DeviceOrientation.portraitUp'],
    ]);
  });
}
