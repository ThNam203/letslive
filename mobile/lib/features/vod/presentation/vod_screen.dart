/// The video screen, `/users/:userId/vods/:vodId` (web
/// `app/(main)/users/[userId]/vods/[vodId]`): the player, title, uploader,
/// like/dislike, description, the uploader's other videos and the comments.
///
/// - A video that does not exist, was deleted or is private to someone else
///   shows "not found" with a way back (Review Focus 5).
/// - Like web, the route's `:userId` is not checked against the video's
///   owner: the uploader row, "Other streams" and the "Owner" badge use the
///   route's user.
/// - The player pins to the top while the page scrolls, so a tapped
///   timestamp in the description or a comment is seen to work. With the
///   keyboard open it is hidden (but keeps playing) to leave room to type.
/// - One view is counted per opening, once the viewer has watched long
///   enough (`domain/view_threshold.dart`). A refused request is not
///   retried: web retries on every progress tick, which on a phone would
///   hammer a server that keeps saying no.
library;

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/widgets/back_notice.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
import 'package:letslive_mobile/features/vod/domain/view_threshold.dart';
import 'package:letslive_mobile/features/vod/presentation/comments_section.dart';
import 'package:letslive_mobile/features/vod/presentation/other_vods_section.dart';
import 'package:letslive_mobile/features/vod/presentation/timestamped_text.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_details.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_player_view.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';
import 'package:letslive_mobile/features/watch/domain/player_controller.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

/// The screen of video [vodId] uploaded by [userId].
class VodScreen extends ConsumerWidget {
  const VodScreen({required this.userId, required this.vodId, super.key});

  /// The uploader, from the route.
  final String userId;

  /// The video, from the route.
  final String vodId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final vod = ref.watch(vodProvider(vodId));
    // Asked together with the video so the uploader's name is likely there
    // when the player starts. The player never waits for it: without it the
    // background-playback notification just has no author line.
    ref.watch(publicUserProvider(userId));
    final Widget body;
    if (vod.hasError && !vod.isLoading) {
      body = Center(
        child: FeedError(
          error: vod.error!,
          onRetry: () => ref.invalidate(vodProvider(vodId)),
        ),
      );
    } else if (!vod.hasValue) {
      body = const Center(child: CircularProgressIndicator());
    } else {
      final found = vod.requireValue;
      body = found == null
          ? BackNotice(
              icon: Icons.videocam_off_outlined,
              message: l10n.api_response_res_err_vod_not_found,
            )
          : _VodBody(vod: found, userId: userId);
    }
    return Scaffold(
      appBar: AppBar(
        title: Text(
          vod.value?.title ?? '',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ),
      body: body,
    );
  }
}

/// The loaded video. Owns the player, so a tapped timestamp can seek it and
/// playback can be counted as a view.
class _VodBody extends ConsumerStatefulWidget {
  const _VodBody({required this.vod, required this.userId});

  /// The loaded video.
  final Vod vod;

  /// The uploader, from the route.
  final String userId;

  @override
  ConsumerState<_VodBody> createState() => _VodBodyState();
}

class _VodBodyState extends ConsumerState<_VodBody> {
  /// Null when the video has no playback URL yet (still processing).
  PlayerController? _player;

  /// Read once: `ref` is not usable in [dispose].
  late final ScreenWake _screenWake;

  bool _awake = false;

  /// A view was (or is being) registered for this opening.
  bool _viewRegistered = false;

  @override
  void initState() {
    super.initState();
    _screenWake = ref.read(screenWakeProvider);
    _player = _createPlayer();
  }

  PlayerController? _createPlayer() {
    final url = widget.vod.playbackUrl;
    if (url == null) return null;
    final player = ref.read(playerFactoryProvider)(
      url: url,
      title: widget.vod.title,
      author: ref.read(publicUserProvider(widget.userId)).value?.username ?? '',
      isLive: false,
    );
    player.value.addListener(_onPlayerChanged);
    return player;
  }

  void _releasePlayer() {
    final player = _player;
    if (player == null) return;
    player.value.removeListener(_onPlayerChanged);
    player.dispose();
    _setAwake(false);
  }

  void _onPlayerChanged() {
    final value = _player?.value.value;
    if (value == null) return;
    _setAwake(value.isPlaying && !value.hasError);
    _countView(value.position);
  }

  void _setAwake(bool awake) {
    if (awake == _awake) return;
    _awake = awake;
    unawaited(_screenWake(awake));
  }

  void _countView(Duration position) {
    if (_viewRegistered) return;
    final seconds = position.inMilliseconds / 1000;
    if (!shouldRegisterView(
      playedSeconds: seconds,
      duration: widget.vod.duration,
    )) {
      return;
    }
    _viewRegistered = true;
    unawaited(
      ref
          .read(vodRepositoryProvider)
          .registerView(widget.vod.id, seconds.floor()),
    );
  }

  /// Starts a fresh player after an error, back where the viewer was (the
  /// player holds the seek until the video can play).
  void _retry() => setState(() {
    final resumeAt = _player?.value.value.position ?? Duration.zero;
    _releasePlayer();
    final player = _player = _createPlayer();
    if (player != null && resumeAt > Duration.zero) {
      unawaited(player.seekTo(resumeAt));
    }
  });

  /// Jumps to a tapped timestamp and plays, like web's `seekTo`.
  void _seekTo(int seconds) {
    final player = _player;
    if (player == null) return;
    unawaited(player.seekTo(Duration(seconds: seconds)));
    unawaited(player.play());
  }

  @override
  void dispose() {
    _releasePlayer();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final vod = widget.vod;
    final player = _player;
    final keyboardOpen = MediaQuery.viewInsetsOf(context).bottom > 0;
    return Column(
      children: [
        Visibility(
          visible: !keyboardOpen,
          maintainState: true,
          child: player == null
              ? const _NotPlayable()
              : VodPlayerView(player: player, onRetry: _retry),
        ),
        Expanded(
          child: VodPlayback(
            seekTo: _seekTo,
            duration: vod.duration,
            child: ListView(
              children: [
                const SizedBox(height: 12),
                VodDetails(vod: vod, ownerId: widget.userId),
                const SizedBox(height: 8),
                OtherVodsSection(userId: widget.userId, currentVodId: vod.id),
                CommentsSection(vodId: vod.id, ownerId: widget.userId),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

/// The player's place for a video that cannot be played (no playback URL).
class _NotPlayable extends StatelessWidget {
  const _NotPlayable();

  @override
  Widget build(BuildContext context) => AspectRatio(
    aspectRatio: 16 / 9,
    child: ColoredBox(
      color: Colors.black,
      child: Center(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Text(
            AppLocalizations.of(context).users_channel_playback_error,
            textAlign: TextAlign.center,
            style: const TextStyle(color: Colors.white),
          ),
        ),
      ),
    ),
  );
}
