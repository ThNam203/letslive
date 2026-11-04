import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:letslive_mobile/features/watch/domain/live_player_controller.dart';

/// A [LivePlayerController] without a native player: starts out playing,
/// records what the screen asked for, and lets a test [emit] any state.
class FakeLivePlayer implements LivePlayerController {
  FakeLivePlayer({
    required this.url,
    required this.title,
    required this.author,
    List<VideoQuality> qualities = const [],
  }) : _value = ValueNotifier(
         LivePlayerValue(
           isPlaying: true,
           isBuffering: false,
           qualities: qualities,
         ),
       );

  final String url;
  final String title;
  final String author;
  final ValueNotifier<LivePlayerValue> _value;

  /// Every quality the screen selected, in order.
  final List<VideoQuality> selected = [];
  int pipRequests = 0;
  bool disposed = false;

  /// Replaces the playback state, as the native player would.
  void emit(LivePlayerValue value) => _value.value = value;

  @override
  ValueListenable<LivePlayerValue> get value => _value;

  @override
  Widget buildVideo(BuildContext context) =>
      const SizedBox.expand(key: ValueKey('fake-video'));

  @override
  Future<void> play() async => emit(
    LivePlayerValue(
      isPlaying: true,
      isBuffering: false,
      qualities: _value.value.qualities,
      quality: _value.value.quality,
    ),
  );

  @override
  Future<void> pause() async => emit(
    LivePlayerValue(
      isBuffering: false,
      qualities: _value.value.qualities,
      quality: _value.value.quality,
    ),
  );

  @override
  Future<void> selectQuality(VideoQuality quality) async {
    selected.add(quality);
    emit(
      LivePlayerValue(
        isPlaying: true,
        isBuffering: false,
        qualities: _value.value.qualities,
        quality: quality,
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

/// The player factory for tests: builds [FakeLivePlayer]s with [qualities]
/// and keeps them in [created].
class FakeLivePlayers {
  FakeLivePlayers({this.qualities = const []});

  final List<VideoQuality> qualities;
  final List<FakeLivePlayer> created = [];

  /// The most recently created player.
  FakeLivePlayer get last => created.last;

  /// A `LivePlayerFactory`.
  LivePlayerController create({
    required String url,
    required String title,
    required String author,
  }) {
    final player = FakeLivePlayer(
      url: url,
      title: title,
      author: author,
      qualities: qualities,
    );
    created.add(player);
    return player;
  }
}
