import 'package:better_player_plus/better_player_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:letslive_mobile/features/watch/domain/live_player_controller.dart';

/// The [LivePlayerController] the app uses: `better_player_plus` (ExoPlayer /
/// AVPlayer). See `mobile/README.md` ("Video player") for why this package.
///
/// - Autoplays the stream with the package's own controls hidden; the
///   screen draws its own overlay.
/// - Keeps playing in the background and in picture-in-picture: lifecycle
///   pausing is off and the media notification / lock screen controls are
///   on (that also keeps an Android foreground service alive).
/// - Qualities are the master playlist variants the package parsed.
class BetterLivePlayerController implements LivePlayerController {
  BetterLivePlayerController({
    required String url,
    required String title,
    required String author,
  }) {
    _player = BetterPlayerController(
      const BetterPlayerConfiguration(
        autoPlay: true,
        aspectRatio: 16 / 9,
        fit: BoxFit.contain,
        handleLifecycle: false,
        // Keep the screen on while watching.
        allowedScreenSleep: false,
        controlsConfiguration: BetterPlayerControlsConfiguration(
          showControls: false,
        ),
      ),
      betterPlayerDataSource: BetterPlayerDataSource.network(
        url,
        liveStream: true,
        videoFormat: BetterPlayerVideoFormat.hls,
        useAsmsTracks: true,
        notificationConfiguration: BetterPlayerNotificationConfiguration(
          showNotification: true,
          title: title,
          author: author,
          // The package opens `<applicationId>.<activityName>`; flavors add
          // an applicationId suffix, so the manifest declares this alias
          // under every applicationId.
          activityName: _notificationActivity,
        ),
      ),
    );
    _player.addEventsListener(_onEvent);
  }

  /// `activity-alias` in `android/app/src/main/AndroidManifest.xml`.
  static const String _notificationActivity = 'PlaybackActivity';

  late final BetterPlayerController _player;
  final GlobalKey _videoKey = GlobalKey();
  final ValueNotifier<LivePlayerValue> _value = ValueNotifier(
    const LivePlayerValue(),
  );
  bool _hasEnded = false;
  bool _hasError = false;
  VideoQuality _quality = VideoQuality.auto;

  @override
  ValueListenable<LivePlayerValue> get value => _value;

  void _onEvent(BetterPlayerEvent event) {
    switch (event.betterPlayerEventType) {
      case BetterPlayerEventType.finished:
        _hasEnded = true;
      case BetterPlayerEventType.exception:
        _hasError = true;
      default:
        break;
    }
    _sync();
  }

  /// Copies the native player's state into [value].
  void _sync() {
    final native = _player.videoPlayerController?.value;
    final ready = native != null && native.initialized;
    _value.value = LivePlayerValue(
      isPlaying: native?.isPlaying ?? false,
      isBuffering: !ready || native.isBuffering,
      hasEnded: _hasEnded,
      hasError: _hasError || (native?.hasError ?? false),
      qualities: qualitiesOf(_player.betterPlayerAsmsTracks),
      quality: _quality,
    );
  }

  @override
  Widget buildVideo(BuildContext context) =>
      BetterPlayer(key: _videoKey, controller: _player);

  @override
  Future<void> play() => _player.play();

  @override
  Future<void> pause() => _player.pause();

  @override
  Future<void> selectQuality(VideoQuality quality) async {
    final track = quality.isAuto
        ? BetterPlayerAsmsTrack.defaultTrack()
        : _player.betterPlayerAsmsTracks
              .where((t) => t.height == quality.height)
              .firstOrNull;
    // A quality from [value] always has a track; nothing to do otherwise.
    if (track == null) return;
    await _player.setTrack(track);
    _quality = quality;
    _sync();
  }

  @override
  Future<void> enterPictureInPicture() async {
    // The package logs and does nothing where PiP is unsupported.
    await _player.enablePictureInPicture(_videoKey);
  }

  @override
  void dispose() {
    _player
      ..removeEventsListener(_onEvent)
      ..dispose(forceDispose: true);
    _value.dispose();
  }
}

/// The selectable variants of [tracks], like web's quality list: one per
/// height in master playlist order (web picks the first variant of a
/// height). Tracks without a height (audio-only, the package's "default"
/// placeholder) are left out.
@visibleForTesting
List<VideoQuality> qualitiesOf(List<BetterPlayerAsmsTrack> tracks) {
  final seen = <int>{};
  return [
    for (final track in tracks)
      if (track.height case final height? when height > 0 && seen.add(height))
        VideoQuality(height),
  ];
}
