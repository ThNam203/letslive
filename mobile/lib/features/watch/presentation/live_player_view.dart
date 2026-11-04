import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/watch/domain/live_player_controller.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

/// A 16:9 live player: the video plus a small overlay (web's live player
/// controls that make sense on a phone): live badge, play/pause, quality
/// menu (Auto + the stream's variants) and picture-in-picture.
///
/// Owns its [LivePlayerController] (from [livePlayerFactoryProvider]) and
/// disposes it with the widget. When playback ends or fails it calls
/// [onStopped] once, so the screen can ask the server whether the stream is
/// still live; if it is, the overlay offers a retry that starts a fresh
/// player.
class LivePlayerView extends ConsumerStatefulWidget {
  const LivePlayerView({
    required this.url,
    required this.title,
    required this.author,
    required this.onStopped,
    super.key,
  });

  /// HLS master playlist.
  final String url;

  /// Stream title, for the background-playback notification.
  final String title;

  /// Streamer name, for the background-playback notification.
  final String author;

  /// Playback ended or failed.
  final VoidCallback onStopped;

  @override
  ConsumerState<LivePlayerView> createState() => _LivePlayerViewState();
}

class _LivePlayerViewState extends ConsumerState<LivePlayerView> {
  late LivePlayerController _player;

  /// Whether [LivePlayerView.onStopped] was called for the current player.
  bool _reportedStop = false;

  @override
  void initState() {
    super.initState();
    _player = _createPlayer();
  }

  LivePlayerController _createPlayer() {
    final player = ref.read(livePlayerFactoryProvider)(
      url: widget.url,
      title: widget.title,
      author: widget.author,
    );
    player.value.addListener(_onPlayerChanged);
    return player;
  }

  void _releasePlayer() {
    _player.value.removeListener(_onPlayerChanged);
    _player.dispose();
  }

  void _onPlayerChanged() {
    final value = _player.value.value;
    if (_reportedStop || !(value.hasEnded || value.hasError)) return;
    _reportedStop = true;
    widget.onStopped();
  }

  void _retry() => setState(() {
    _releasePlayer();
    _reportedStop = false;
    _player = _createPlayer();
  });

  @override
  void dispose() {
    _releasePlayer();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => AspectRatio(
    aspectRatio: 16 / 9,
    child: ColoredBox(
      color: Colors.black,
      child: Stack(
        fit: StackFit.expand,
        children: [
          _player.buildVideo(context),
          ValueListenableBuilder<LivePlayerValue>(
            valueListenable: _player.value,
            builder: (context, value, _) => value.hasError || value.hasEnded
                ? _PlaybackError(onRetry: _retry)
                : _Controls(player: _player, value: value),
          ),
        ],
      ),
    ),
  );
}

/// Spinner while buffering, the control bar at the bottom.
class _Controls extends StatelessWidget {
  const _Controls({required this.player, required this.value});

  final LivePlayerController player;
  final LivePlayerValue value;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Stack(
      children: [
        if (value.isBuffering)
          const Center(child: CircularProgressIndicator(color: Colors.white)),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: ColoredBox(
            color: Colors.black45,
            child: IconTheme(
              data: const IconThemeData(color: Colors.white),
              child: Row(
                children: [
                  const SizedBox(width: 8),
                  _LiveBadge(l10n.users_live),
                  const Spacer(),
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
                  _QualityMenu(player: player, value: value),
                  IconButton(
                    tooltip: l10n.users_player_picture_in_picture,
                    icon: const Icon(Icons.picture_in_picture_alt),
                    onPressed: player.enterPictureInPicture,
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

/// "Auto" plus one entry per variant, the one in use checked.
class _QualityMenu extends StatelessWidget {
  const _QualityMenu({required this.player, required this.value});

  final LivePlayerController player;
  final LivePlayerValue value;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return PopupMenuButton<VideoQuality>(
      key: const ValueKey('player-quality'),
      tooltip: l10n.users_player_quality,
      icon: const Icon(Icons.settings),
      onSelected: player.selectQuality,
      itemBuilder: (context) => [
        for (final quality in [VideoQuality.auto, ...value.qualities])
          CheckedPopupMenuItem(
            value: quality,
            checked: quality == value.quality,
            child: Text(
              quality.isAuto
                  ? l10n.users_player_quality_auto
                  : '${quality.height}p',
            ),
          ),
      ],
    );
  }
}

class _LiveBadge extends StatelessWidget {
  const _LiveBadge(this.label);

  final String label;

  @override
  Widget build(BuildContext context) => DecoratedBox(
    decoration: BoxDecoration(
      color: Colors.red.shade600,
      borderRadius: BorderRadius.circular(4),
    ),
    child: Padding(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelSmall
            ?.copyWith(color: Colors.white),
      ),
    ),
  );
}

/// Playback failed (or the playlist ended) while the server still says the
/// stream is live.
class _PlaybackError extends StatelessWidget {
  const _PlaybackError({required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              l10n.users_channel_playback_error,
              textAlign: TextAlign.center,
              style: const TextStyle(color: Colors.white),
            ),
            const SizedBox(height: 8),
            FilledButton(
              onPressed: onRetry,
              child: Text(l10n.users_channel_retry),
            ),
          ],
        ),
      ),
    );
  }
}
