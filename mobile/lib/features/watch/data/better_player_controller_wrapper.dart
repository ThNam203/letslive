import 'dart:async';

import 'package:better_player_plus/better_player_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';

/// The [PlayerController] the app uses: `better_player_plus` (ExoPlayer /
/// AVPlayer). See `mobile/README.md` ("Video player") for why this package.
///
/// - Autoplays the stream with the package's own controls hidden; the
///   screen draws its own overlay.
/// - Keeps playing in the background and in picture-in-picture: lifecycle
///   pausing is off and the media notification / lock screen controls are
///   on (that also keeps an Android foreground service alive).
/// - Qualities are the master playlist variants the package parsed.
/// - Every call into the package runs in [playerErrorZone], so errors from
///   the package's own timers turn into [PlayerValue.hasError] instead
///   of escaping as uncaught errors.
class BetterPlayerControllerWrapper implements PlayerController {
  BetterPlayerControllerWrapper({
    required String url,
    required String title,
    required String author,
    bool isLive = true,
  }) {
    _zone = playerErrorZone(_onUncaughtError);
    // Built (and so autoplayed) inside the zone: the position timer the
    // package starts on play inherits it.
    _zone.run(() {
      _player = BetterPlayerController(
        const BetterPlayerConfiguration(
          autoPlay: true,
          aspectRatio: 16 / 9,
          fit: BoxFit.contain,
          handleLifecycle: false,
          // Only applies on the package's fullscreen route (which Android
          // picture-in-picture uses); inline, `LivePlayerView` holds the
          // wakelock through `screenWakeProvider`.
          allowedScreenSleep: false,
          controlsConfiguration: BetterPlayerControlsConfiguration(
            showControls: false,
          ),
        ),
        betterPlayerDataSource: BetterPlayerDataSource.network(
          url,
          liveStream: isLive,
          videoFormat: BetterPlayerVideoFormat.hls,
          useAsmsTracks: true,
          notificationConfiguration: BetterPlayerNotificationConfiguration(
            showNotification: true,
            title: title,
            author: author,
            // The package opens `<applicationId>.<activityName>`; flavors
            // add an applicationId suffix, so the manifest declares this
            // alias under every applicationId.
            activityName: _notificationActivity,
          ),
        ),
      );
      _player.addEventsListener(_onEvent);
    });
  }

  /// `activity-alias` in `android/app/src/main/AndroidManifest.xml`.
  static const String _notificationActivity = 'PlaybackActivity';

  late final Zone _zone;
  late final BetterPlayerController _player;
  final GlobalKey _videoKey = GlobalKey();
  final ValueNotifier<PlayerValue> _value = ValueNotifier(const PlayerValue());
  bool _hasEnded = false;
  bool _hasError = false;
  bool _disposed = false;
  VideoQuality _quality = VideoQuality.auto;

  /// A seek asked before the video could play; applied on `initialized`.
  Duration? _pendingSeek;

  @override
  ValueListenable<PlayerValue> get value => _value;

  void _onEvent(BetterPlayerEvent event) {
    switch (event.betterPlayerEventType) {
      case BetterPlayerEventType.initialized:
        final pending = _pendingSeek;
        _pendingSeek = null;
        if (pending != null) unawaited(seekTo(pending));
      case BetterPlayerEventType.finished:
        _hasEnded = true;
      case BetterPlayerEventType.exception:
        _hasError = true;
      default:
        break;
    }
    _sync();
  }

  /// An uncaught package error (other than the ignored `RangeError`): log
  /// it like any Flutter error and show the playback as failed.
  void _onUncaughtError(Object error, StackTrace stack) {
    FlutterError.reportError(
      FlutterErrorDetails(exception: error, stack: stack, library: 'player'),
    );
    _hasError = true;
    _sync();
  }

  /// Copies the native player's state into [value].
  void _sync() {
    // Package events and awaited calls can finish after [dispose].
    if (_disposed) return;
    final native = _player.videoPlayerController?.value;
    final ready = native != null && native.initialized;
    _value.value = PlayerValue(
      isPlaying: native?.isPlaying ?? false,
      isBuffering: !ready || native.isBuffering,
      hasEnded: _hasEnded,
      hasError: _hasError || (native?.hasError ?? false),
      qualities: qualitiesOf(_player.betterPlayerAsmsTracks),
      quality: _quality,
      position: native?.position ?? Duration.zero,
      duration: native?.duration ?? Duration.zero,
    );
  }

  @override
  Future<void> seekTo(Duration position) async {
    final native = _player.videoPlayerController?.value;
    if (native == null || !native.initialized) {
      _pendingSeek = position;
      return;
    }
    await _zone.run(() => _player.seekTo(position));
    _sync();
  }

  @override
  Widget buildVideo(BuildContext context) =>
      BetterPlayer(key: _videoKey, controller: _player);

  @override
  Future<void> play() => _zone.run(_player.play);

  @override
  Future<void> pause() => _zone.run(_player.pause);

  @override
  Future<void> selectQuality(VideoQuality quality) async {
    final track = quality.isAuto
        ? BetterPlayerAsmsTrack.defaultTrack()
        : _player.betterPlayerAsmsTracks
              .where((t) => t.height == quality.height)
              .firstOrNull;
    // A quality from [value] always has a track; nothing to do otherwise.
    if (track == null) return;
    await _zone.run(() => _player.setTrack(track));
    _quality = quality;
    _sync();
  }

  @override
  Future<void> enterPictureInPicture() async {
    // The package throws before the stream is set up, and logs and does
    // nothing where PiP is unsupported.
    if (_player.videoPlayerController == null) return;
    await _zone.run(() => _player.enablePictureInPicture(_videoKey));
  }

  @override
  void dispose() {
    _disposed = true;
    _player
      ..removeEventsListener(_onEvent)
      ..dispose(forceDispose: true);
    _value.dispose();
  }
}

/// A zone that sends uncaught errors of the code run in it, including
/// timers and futures that code starts, to [onError].
///
/// `RangeError` is ignored: the package's 300 ms position timer throws it on
/// Android when a live window has no start time (`getAbsolutePosition` gets
/// an overflowed timestamp). Native playback is unaffected; only the
/// package's position update for that tick is lost.
@visibleForTesting
Zone playerErrorZone(void Function(Object error, StackTrace stack) onError) {
  late final Zone zone;
  runZonedGuarded(() => zone = Zone.current, (error, stack) {
    if (error is RangeError) return;
    onError(error, stack);
  });
  return zone;
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
