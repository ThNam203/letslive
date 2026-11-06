import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:letslive_mobile/features/watch/presentation/live_player_view.dart';
import 'package:letslive_mobile/features/watch/presentation/player_chrome.dart';

/// How far the skip buttons jump (web `SKIP_SECONDS`).
const Duration skipStep = Duration(seconds: 10);

/// A 16:9 player for a recorded video: the video plus a control bar with a
/// seek slider, play/pause, 10 second skips, the time, the settings menu
/// (picture-in-picture and quality) and full screen. Tapping the video pauses or plays
/// it; the bar fades out while playing ([PlayerChrome]). The [player] is owned by the screen (it needs it to
/// seek from timestamps and to count the view); this view only draws it.
///
/// When playback fails it shows [onRetry] instead of the controls.
class VodPlayerView extends StatelessWidget {
  const VodPlayerView({required this.player, required this.onRetry, super.key});

  /// The player to show and control.
  final PlayerController player;

  /// Starts a fresh player after a playback error.
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => AspectRatio(
    aspectRatio: 16 / 9,
    child: ColoredBox(
      color: Colors.black,
      child: Stack(
        fit: StackFit.expand,
        children: [
          player.buildVideo(context),
          ValueListenableBuilder<PlayerValue>(
            valueListenable: player.value,
            builder: (context, value, _) => value.hasError
                ? PlayerError(onRetry: onRetry)
                : _Controls(player: player, value: value),
          ),
        ],
      ),
    ),
  );
}

class _Controls extends StatefulWidget {
  const _Controls({required this.player, required this.value});

  final PlayerController player;
  final PlayerValue value;

  @override
  State<_Controls> createState() => _ControlsState();
}

class _ControlsState extends State<_Controls> {
  /// Slider position (seconds) while the finger is on it; null otherwise.
  double? _dragging;

  /// Moves by [delta] from the current position, within the video.
  void _skip(Duration delta) {
    final value = widget.value;
    final target = value.position + delta;
    final end = value.duration;
    final clamped = target < Duration.zero
        ? Duration.zero
        : (end > Duration.zero && target > end ? end : target);
    widget.player.seekTo(clamped);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final value = widget.value;
    final player = widget.player;
    final total = value.duration.inSeconds;
    final shown = _dragging ?? value.position.inSeconds.toDouble();
    final time = '${formatDuration(shown.round())} / ${formatDuration(total)}';
    return PlayerChrome(
      player: player,
      value: value,
      bar: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (total > 0)
            Slider(
              value: shown.clamp(0, total.toDouble()),
              max: total.toDouble(),
              onChanged: (v) => setState(() => _dragging = v),
              onChangeEnd: (v) {
                setState(() => _dragging = null);
                player.seekTo(Duration(seconds: v.round()));
              },
            ),
          Row(
            children: [
              PlayPauseButton(player: player, value: value),
              IconButton(
                tooltip: l10n.accessibility_seek_back_10,
                icon: const Icon(Icons.replay_10),
                onPressed: () => _skip(-skipStep),
              ),
              IconButton(
                tooltip: l10n.accessibility_seek_forward_10,
                icon: const Icon(Icons.forward_10),
                onPressed: () => _skip(skipStep),
              ),
              Expanded(
                child: Text(
                  time,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.labelSmall
                      ?.copyWith(color: Colors.white),
                ),
              ),
              PlayerSettingsMenu(player: player, value: value),
              const FullscreenButton(),
            ],
          ),
        ],
      ),
    );
  }
}
