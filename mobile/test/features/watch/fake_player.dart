import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';

/// A [PlayerController] without a native player: starts out playing,
/// records what the screen asked for, and lets a test [emit] any state.
class FakePlayer implements PlayerController {
  FakePlayer({
    required this.url,
    required this.title,
    required this.author,
    this.isLive = true,
    List<VideoQuality> qualities = const [],
  }) : _value = ValueNotifier(
         PlayerValue(isPlaying: true, isBuffering: false, qualities: qualities),
       );

  final String url;
  final String title;
  final String author;

  /// Whether the screen asked for a live stream (false = recorded video).
  final bool isLive;
  final ValueNotifier<PlayerValue> _value;

  /// Every position the screen sought to, in order.
  final List<Duration> seeks = [];

  /// Every quality the screen selected, in order.
  final List<VideoQuality> selected = [];
  int pipRequests = 0;
  bool disposed = false;

  /// Replaces the playback state, as the native player would.
  void emit(PlayerValue value) => _value.value = value;

  @override
  ValueListenable<PlayerValue> get value => _value;

  @override
  Widget buildVideo(BuildContext context) =>
      const SizedBox.expand(key: ValueKey('fake-video'));

  @override
  Future<void> play() async => emit(
    PlayerValue(
      isPlaying: true,
      isBuffering: false,
      qualities: _value.value.qualities,
      quality: _value.value.quality,
    ),
  );

  @override
  Future<void> pause() async => emit(
    PlayerValue(
      isBuffering: false,
      qualities: _value.value.qualities,
      quality: _value.value.quality,
    ),
  );

  @override
  Future<void> selectQuality(VideoQuality quality) async {
    selected.add(quality);
    emit(
      PlayerValue(
        isPlaying: true,
        isBuffering: false,
        qualities: _value.value.qualities,
        quality: quality,
      ),
    );
  }

  @override
  Future<void> seekTo(Duration position) async {
    seeks.add(position);
    final old = _value.value;
    emit(
      PlayerValue(
        isPlaying: old.isPlaying,
        isBuffering: false,
        qualities: old.qualities,
        quality: old.quality,
        position: position,
        duration: old.duration,
      ),
    );
  }

  @override
  Future<void> enterPictureInPicture() async => pipRequests++;

  @override
  void dispose() {
    disposed = true;
    _value.dispose();
  }
}

/// A `ScreenWake` for tests: records every request instead of calling the
/// wakelock plugin.
class FakeScreenWake {
  /// Every request in order: true = keep the screen on.
  final List<bool> calls = [];

  /// Whether the screen is currently kept on.
  bool get isAwake => calls.lastOrNull ?? false;

  /// A `ScreenWake`.
  Future<void> call(bool awake) async => calls.add(awake);
}

/// The player factory for tests: builds [FakePlayer]s with [qualities]
/// and keeps them in [created].
class FakePlayers {
  FakePlayers({this.qualities = const []});

  final List<VideoQuality> qualities;
  final List<FakePlayer> created = [];

  /// The most recently created player.
  FakePlayer get last => created.last;

  /// A `PlayerFactory`.
  PlayerController create({
    required String url,
    required String title,
    required String author,
    bool isLive = true,
  }) {
    final player = FakePlayer(
      url: url,
      title: title,
      author: author,
      isLive: isLive,
      qualities: qualities,
    );
    created.add(player);
    return player;
  }
}
