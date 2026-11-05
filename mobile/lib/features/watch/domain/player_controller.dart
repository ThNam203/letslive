/// The boundary between the watch screens and the video player package.
///
/// Screens only talk to a [LivePlayerController]; the app builds the
/// `better_player_plus` one (`BetterLivePlayerController`) and widget tests a
/// fake. Kept small on purpose: what the live screen and the VOD screen
/// show and do. It plays live streams and recorded videos alike.
library;

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';

/// One HLS variant the viewer can pick, named by its height like web's
/// player ("720p"). [auto] lets the player adapt to the bandwidth.
@immutable
class VideoQuality {
  const VideoQuality(this.height);

  /// Adaptive bitrate: the player picks the variant (web's "Auto").
  static const VideoQuality auto = VideoQuality(0);

  /// Vertical resolution in pixels; 0 for [auto].
  final int height;

  /// Whether this is [auto] (adaptive) rather than a fixed variant.
  bool get isAuto => height == 0;

  @override
  bool operator ==(Object other) =>
      other is VideoQuality && other.height == height;

  @override
  int get hashCode => height.hashCode;

  @override
  String toString() =>
      isAuto ? 'VideoQuality.auto' : 'VideoQuality(${height}p)';
}

/// What the player overlay shows about playback.
@immutable
class LivePlayerValue {
  const LivePlayerValue({
    this.isPlaying = false,
    this.isBuffering = true,
    this.hasEnded = false,
    this.hasError = false,
    this.qualities = const [],
    this.quality = VideoQuality.auto,
    this.position = Duration.zero,
    this.duration = Duration.zero,
  });

  /// Playing (or about to, once buffered) rather than paused by the viewer.
  final bool isPlaying;

  /// Waiting for data (start-up or a stall).
  final bool isBuffering;

  /// The stream's playlist ended (the streamer stopped).
  final bool hasEnded;

  /// Playback failed (network, 404 playlist, decoder).
  final bool hasError;

  /// The stream's variants in master playlist order, without duplicates;
  /// empty until the master playlist was read.
  final List<VideoQuality> qualities;

  /// The variant in use; [VideoQuality.auto] by default.
  final VideoQuality quality;

  /// Playback position (recorded videos; the live screen ignores it).
  final Duration position;

  /// Length of the video, zero until known (recorded videos only).
  final Duration duration;
}

/// Plays one HLS stream. Created per stream by [LivePlayerFactory] and
/// disposed by the widget that created it.
abstract interface class LivePlayerController {
  /// Current playback state; listeners are told about every change.
  ValueListenable<LivePlayerValue> get value;

  /// The video surface (16:9, letterboxed).
  Widget buildVideo(BuildContext context);

  /// Resumes playback after [pause].
  Future<void> play();

  /// Pauses playback; the stream keeps going on the server.
  Future<void> pause();

  /// Jumps to [position] in a recorded video. Asked before the video can
  /// play, it is held and applied once it can (like web's pending seek).
  Future<void> seekTo(Duration position);

  /// Switches to [quality]; [VideoQuality.auto] goes back to adaptive.
  Future<void> selectQuality(VideoQuality quality);

  /// Moves the video into a floating picture-in-picture window. Does nothing
  /// where the device does not support it.
  Future<void> enterPictureInPicture();

  /// Stops playback and releases the native player.
  void dispose();
}

/// Builds the player for the stream at [url]. [title] and [author] label the
/// background-playback notification / lock screen controls.
///
/// [isLive] is true for a stream (the default) and false for a recorded
/// video (VOD), which can seek and reports position and duration. The name
/// `LivePlayerController` is kept for both: one wrapper, one fake.
typedef LivePlayerFactory = LivePlayerController Function({
  required String url,
  required String title,
  required String author,
  bool isLive,
});
