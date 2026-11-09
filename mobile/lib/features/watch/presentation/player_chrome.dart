/// What the live and the VOD player share around their control bar: tap the
/// video to pause or play, controls that fade out after
/// [controlsHideDelay] without a touch while the video plays, and the
/// full-screen button.
library;

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:letslive_mobile/features/watch/presentation/live_player_view.dart';

/// How long the controls stay after the last touch while the video plays.
const Duration controlsHideDelay = Duration(seconds: 3);

/// How long the controls take to fade in or out.
const Duration _fadeDuration = Duration(milliseconds: 200);

/// The overlay of a player: a tap anywhere on the video pauses or plays
/// [player], and [bar] (the control bar) sits at the bottom.
///
/// The bar shows while the video is paused and for [controlsHideDelay]
/// after the last touch while it plays; then it fades out and lets taps
/// through to the video. Any touch, on the video or on the bar, brings it
/// back, and a finger held down (dragging the seek slider) keeps it up.
class PlayerChrome extends StatefulWidget {
  const PlayerChrome({
    required this.player,
    required this.value,
    required this.bar,
    super.key,
  });

  final PlayerController player;
  final PlayerValue value;
  final Widget bar;

  @override
  State<PlayerChrome> createState() => _PlayerChromeState();
}

class _PlayerChromeState extends State<PlayerChrome> {
  bool _visible = true;
  Timer? _hideTimer;

  /// Fingers on the overlay; the bar does not hide while any is down.
  int _pointers = 0;

  @override
  void initState() {
    super.initState();
    _scheduleHide();
  }

  @override
  void didUpdateWidget(PlayerChrome old) {
    super.didUpdateWidget(old);
    if (old.value.isPlaying == widget.value.isPlaying) return;
    // Paused: keep the controls up. Playing again: count down from now.
    _visible = true;
    _scheduleHide();
  }

  void _scheduleHide() {
    _hideTimer?.cancel();
    if (!widget.value.isPlaying || _pointers > 0) return;
    _hideTimer = Timer(controlsHideDelay, () {
      if (mounted) setState(() => _visible = false);
    });
  }

  void _onPointerDown(PointerDownEvent _) {
    _pointers++;
    _hideTimer?.cancel();
    if (!_visible) setState(() => _visible = true);
  }

  void _onPointerUp(PointerEvent _) {
    if (_pointers > 0) _pointers--;
    _scheduleHide();
  }

  void _togglePlay() {
    final player = widget.player;
    unawaited(widget.value.isPlaying ? player.pause() : player.play());
  }

  @override
  void dispose() {
    _hideTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Listener(
    onPointerDown: _onPointerDown,
    onPointerUp: _onPointerUp,
    onPointerCancel: _onPointerUp,
    child: Stack(
      fit: StackFit.expand,
      children: [
        GestureDetector(
          key: const ValueKey('player-surface'),
          behavior: HitTestBehavior.opaque,
          onTap: _togglePlay,
        ),
        if (widget.value.isBuffering)
          const IgnorePointer(child: PlayerSpinner()),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: IgnorePointer(
            ignoring: !_visible,
            child: AnimatedOpacity(
              key: const ValueKey('player-controls'),
              opacity: _visible ? 1 : 0,
              duration: _fadeDuration,
              child: ColoredBox(
                color: Colors.black45,
                child: IconTheme(
                  data: const IconThemeData(color: Colors.white),
                  child: widget.bar,
                ),
              ),
            ),
          ),
        ),
      ],
    ),
  );
}

/// Enters full screen in portrait, leaves it sideways. Full screen is the
/// landscape layout of the watch screens ([LandscapePlayer]): the button
/// turns the screen, and rotating the phone does the same.
class FullscreenButton extends StatelessWidget {
  const FullscreenButton({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final fullscreen = isLandscape(context);
    return IconButton(
      key: const ValueKey('player-fullscreen'),
      tooltip: fullscreen
          ? l10n.users_player_exit_fullscreen
          : l10n.users_player_fullscreen,
      icon: Icon(fullscreen ? Icons.fullscreen_exit : Icons.fullscreen),
      onPressed: () => unawaited(setPlayerFullscreen(on: !fullscreen)),
    );
  }
}

/// Locks the screen sideways ([on]) or upright. Upright stays locked until
/// the watch screen closes ([resetPlayerOrientation]); otherwise a phone
/// still held sideways would turn straight back.
Future<void> setPlayerFullscreen({required bool on}) =>
    SystemChrome.setPreferredOrientations(
      on
          ? const [
              DeviceOrientation.landscapeLeft,
              DeviceOrientation.landscapeRight,
            ]
          : const [DeviceOrientation.portraitUp],
    );

/// Gives the screen back to the phone's rotation; the watch screens call it
/// when they close.
Future<void> resetPlayerOrientation() =>
    SystemChrome.setPreferredOrientations(const []);
