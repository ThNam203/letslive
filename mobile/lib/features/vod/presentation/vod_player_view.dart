import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:letslive_mobile/features/watch/presentation/live_player_view.dart';

/// How far the skip buttons jump (web `SKIP_SECONDS`).
const Duration skipStep = Duration(seconds: 10);

/// A 16:9 player for a recorded video: the video plus a control bar with a
/// seek slider, play/pause, 10 second skips, the time, the quality menu and
/// picture-in-picture. The [player] is owned by the screen (it needs it to
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
    return Stack(
      children: [
        if (value.isBuffering) const PlayerSpinner(),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: ColoredBox(
            color: Colors.black45,
            child: IconTheme(
              data: const IconThemeData(color: Colors.white),
              child: Column(
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
                      value.isPlaying
                          ? IconButton(
                              tooltip: l10n.users_player_pause,
                              icon: const Icon(Icons.pause),
                              onPressed: player.pause,
                            )
                          : IconButton(
                              tooltip: l10n.users_player_play,
                              icon: const Icon(Icons.play_arrow),
                              onPressed: player.play,
                            ),
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
                      PlayerQualityMenu(player: player, value: value),
                      IconButton(
                        tooltip: l10n.users_player_picture_in_picture,
                        icon: const Icon(Icons.picture_in_picture_alt),
                        onPressed: player.enterPictureInPicture,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}
