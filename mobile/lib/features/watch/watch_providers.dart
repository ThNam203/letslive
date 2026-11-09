/// Riverpod providers for watching a channel live (`/users/:userId`).
///
/// How a stream ending is noticed: the backend publishes no realtime event
/// for it (the realtime service only relays chat, member and notification
/// events), and web simply refetches the live stream when the tab regains
/// focus after its 30s stale time. Mobile re-checks `GET
/// /livestreams?userId=` every [liveRecheckInterval] while the screen is
/// open, and at once when the player reports the end or an error. A check
/// that finds no stream turns a live channel into [ChannelEnded]; one that
/// finds a stream turns an offline channel live.
library;

import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
import 'package:letslive_mobile/features/watch/data/better_player_controller_wrapper.dart';
import 'package:letslive_mobile/features/watch/data/watch_repository.dart';
import 'package:letslive_mobile/features/watch/domain/channel_state.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:wakelock_plus/wakelock_plus.dart';

/// How often an open channel asks whether the stream is still live (web's
/// query stale time).
const Duration liveRecheckInterval = Duration(seconds: 30);

/// Live stream lookups.
final watchRepositoryProvider = Provider<WatchRepository>(
  (ref) => WatchRepository(ref.watch(apiClientProvider)),
);

/// Builds the native player; widget tests override it with a fake.
final playerFactoryProvider = Provider<PlayerFactory>(
  (ref) => BetterPlayerControllerWrapper.new,
);

/// Keeps the screen on (`true`) or lets it sleep again (`false`): the raw
/// wakelock toggle. Widget tests override it with a fake; features use
/// [screenWakeHolderProvider], never this directly.
typedef ScreenWake = Future<void> Function(bool awake);

final screenWakeProvider = Provider<ScreenWake>(
  (ref) =>
      (awake) => WakelockPlus.toggle(enable: awake),
);

/// The wakelock is one global switch, but the live player, the video player
/// and a running upload each need the screen on. This holder keeps the set of
/// owners that want it and flips the switch only when the first owner
/// arrives and the last one leaves, so one feature letting go never puts the
/// screen to sleep under another.
class ScreenWakeHolder {
  ScreenWakeHolder(this._toggle);

  final ScreenWake _toggle;
  final Set<Object> _owners = {};

  /// [owner] wants the screen on. Asking twice is the same as once.
  void acquire(Object owner) {
    final wasEmpty = _owners.isEmpty;
    if (_owners.add(owner) && wasEmpty) unawaited(_toggle(true));
  }

  /// [owner] no longer needs the screen on; the screen may sleep once every
  /// owner has released. Releasing without holding does nothing.
  void release(Object owner) {
    if (_owners.remove(owner) && _owners.isEmpty) unawaited(_toggle(false));
  }
}

/// The shared [ScreenWakeHolder].
final screenWakeHolderProvider = Provider<ScreenWakeHolder>(
  (ref) => ScreenWakeHolder(ref.watch(screenWakeProvider)),
);

/// The channel of one user. `autoDispose`: leaving the screen stops the
/// re-checks, and logout drops it.
final channelProvider = AsyncNotifierProvider.autoDispose
    .family<ChannelNotifier, ChannelState, String>(
      ChannelNotifier.new,
      retry: noRetry,
    );

/// Loads a user's profile and live stream, then keeps the live state fresh
/// (see the library doc).
class ChannelNotifier extends AsyncNotifier<ChannelState> {
  ChannelNotifier(this.userId);

  /// The channel owner.
  final String userId;

  @override
  Future<ChannelState> build() async {
    // The profile comes from the shared user provider, which the profile
    // header watches (follow changes the count there); keep it alive so it
    // is not dropped between this load and the header's first build.
    ref.listen(publicUserProvider(userId), (_, _) {});
    final liveFuture = ref.watch(watchRepositoryProvider).liveOf(userId);
    final PublicUser profile;
    try {
      profile = await ref.read(publicUserProvider(userId).future);
    } on ApiLoadException catch (e) {
      // 404 for an unknown id, 400 for one that is not a UUID.
      final status = e.response.statusCode;
      if (status == 404 || status == 400) return const ChannelNotFound();
      rethrow;
    }
    final live = await liveFuture;
    if (!live.success) throw ApiLoadException(live);

    // Keeps polling in the background too: background audio may be playing,
    // and a live playlist that stops growing may never report its end.
    final timer = Timer.periodic(liveRecheckInterval, (_) => recheck());
    ref.onDispose(timer.cancel);

    final stream = live.data;
    return stream == null
        ? ChannelOffline(profile)
        : ChannelLive(profile, stream);
  }

  /// Asks the server whether [userId] is live now and updates a live or
  /// offline channel. A failed check changes nothing (the next one retries).
  Future<void> recheck() async {
    final before = state.value;
    if (before is! ChannelLive && before is! ChannelOffline) return;
    final live = await ref.read(watchRepositoryProvider).liveOf(userId);
    if (!ref.mounted || !live.success) return;
    final stream = live.data;
    switch (state.value) {
      case ChannelLive(:final user):
        state = AsyncData(
          stream == null ? ChannelEnded(user) : ChannelLive(user, stream),
        );
      case ChannelOffline(:final user):
        if (stream != null) state = AsyncData(ChannelLive(user, stream));
      case ChannelEnded() || ChannelNotFound() || null:
        break;
    }
  }
}
