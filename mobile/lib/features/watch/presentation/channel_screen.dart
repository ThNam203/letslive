/// The channel screen, `/users/:userId` (web `app/(main)/users/[userId]`):
/// the live player and title while the user streams, otherwise an offline
/// notice. A stream that ends while watched turns into an "ended" notice;
/// an unknown user shows "not found". Every notice has a way back.
///
/// The live chat (`chat/`) sits under the player, and under the offline /
/// ended notice. The profile header, tabs, follow button and videos
/// (Task 15) slot in between the player and the chat. How the end of a stream is noticed is
/// described in `watch_providers.dart`.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/widgets/back_notice.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/watch/chat/presentation/chat_panel.dart';
import 'package:letslive_mobile/features/watch/domain/channel_state.dart';
import 'package:letslive_mobile/features/watch/domain/live_stream_url.dart';
import 'package:letslive_mobile/features/watch/presentation/live_player_view.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

/// The channel of [userId].
class ChannelScreen extends ConsumerWidget {
  const ChannelScreen({required this.userId, super.key});

  /// The channel owner, from the route.
  final String userId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    // Read here: the Scaffold below hides the keyboard inset from its body.
    final keyboardOpen = MediaQuery.viewInsetsOf(context).bottom > 0;
    final channel = ref.watch(channelProvider(userId));
    final state = channel.value;
    final user = switch (state) {
      ChannelLive(:final user) ||
      ChannelOffline(:final user) ||
      ChannelEnded(:final user) => user,
      ChannelNotFound() || null => null,
    };

    final error = channel.error;
    final Widget body;
    if (error != null && !channel.isLoading) {
      body = Center(
        child: FeedError(
          error: error,
          onRetry: () => ref.invalidate(channelProvider(userId)),
        ),
      );
    } else if (state == null) {
      body = const Center(child: CircularProgressIndicator());
    } else {
      body = switch (state) {
        ChannelLive(:final user, :final stream) => _LiveBody(
          user: user,
          stream: stream,
          keyboardOpen: keyboardOpen,
          onStopped: ref.read(channelProvider(userId).notifier).recheck,
        ),
        ChannelOffline(:final user) => _NoticeWithChat(
          roomId: user.id,
          keyboardOpen: keyboardOpen,
          notice: BackNotice(
            icon: Icons.videocam_off_outlined,
            title: l10n.users_offline,
            message: l10n.users_channel_offline_notice(user.username),
          ),
        ),
        ChannelEnded(:final user) => _NoticeWithChat(
          roomId: user.id,
          keyboardOpen: keyboardOpen,
          notice: BackNotice(
            icon: Icons.live_tv_outlined,
            message: l10n.users_channel_stream_ended,
          ),
        ),
        ChannelNotFound() => BackNotice(
          icon: Icons.person_off_outlined,
          message: l10n.api_response_res_err_user_not_found,
        ),
      };
    }

    return Scaffold(
      appBar: AppBar(
        title: user == null
            ? null
            : Text(user.username, maxLines: 1, overflow: TextOverflow.ellipsis),
      ),
      body: body,
    );
  }
}

/// Chat height below which the live layout scrolls instead of squeezing the
/// chat: room for a few lines plus the message box (and its suggestions).
const double _minChatHeight = 240;
const double _minCompactChatHeight = 200;

/// The player, the stream title and the streamer's name, then the chat.
///
/// With the keyboard open ([keyboardOpen]) the title block goes and the chat
/// drops its heading. If the chat would still get less than its minimum
/// height the page scrolls (player on top, chat at the minimum) rather than
/// overflow. Player and chat hold [GlobalKey]s so switching between the two
/// layouts keeps playback and the typed text.
class _LiveBody extends ConsumerStatefulWidget {
  const _LiveBody({
    required this.user,
    required this.stream,
    required this.keyboardOpen,
    required this.onStopped,
  });

  final PublicUser user;
  final Livestream stream;
  final bool keyboardOpen;
  final Future<void> Function() onStopped;

  @override
  ConsumerState<_LiveBody> createState() => _LiveBodyState();
}

class _LiveBodyState extends ConsumerState<_LiveBody> {
  var _playerKey = GlobalKey();
  final _chatKey = GlobalKey();

  @override
  void didUpdateWidget(_LiveBody old) {
    super.didUpdateWidget(old);
    // A new stream (the user restarted) gets a new player.
    if (old.stream.id != widget.stream.id) _playerKey = GlobalKey();
  }

  /// An upper bound of the title block's height at the current text scale.
  double _infoHeight(BuildContext context) {
    final scaler = MediaQuery.textScalerOf(context);
    return 33 + scaler.scale(16) * 1.5 * 2 + scaler.scale(14) * 1.5;
  }

  @override
  Widget build(BuildContext context) {
    final user = widget.user;
    final stream = widget.stream;
    final keyboard = widget.keyboardOpen;
    final theme = Theme.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final player = LivePlayerView(
      key: _playerKey,
      url: liveStreamUrl(apiBaseUrl, stream.id),
      title: stream.title,
      author: user.username,
      onStopped: widget.onStopped,
    );
    final info = Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            stream.title,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.titleMedium,
          ),
          Text(
            user.username,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
    );
    // The room id is the streamer's user id.
    final chat = ChatPanel(key: _chatKey, roomId: user.id, compact: keyboard);
    final minChat = keyboard ? _minCompactChatHeight : _minChatHeight;

    return LayoutBuilder(
      builder: (context, box) {
        final used =
            box.maxWidth * 9 / 16 + (keyboard ? 0 : _infoHeight(context));
        if (box.maxHeight - used >= minChat) {
          return Column(
            children: [
              player,
              if (!keyboard) info,
              const Divider(height: 1),
              Expanded(child: chat),
            ],
          );
        }
        return SingleChildScrollView(
          child: Column(
            children: [
              player,
              if (!keyboard) info,
              const Divider(height: 1),
              SizedBox(height: minChat, child: chat),
            ],
          ),
        );
      },
    );
  }
}

/// The offline / ended [notice] above the room's chat, which web also shows
/// when nobody is streaming. With the keyboard open only the chat remains.
class _NoticeWithChat extends StatelessWidget {
  const _NoticeWithChat({
    required this.roomId,
    required this.notice,
    required this.keyboardOpen,
  });

  final String roomId;
  final Widget notice;
  final bool keyboardOpen;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      if (!keyboardOpen) ...[Expanded(child: notice), const Divider(height: 1)],
      // Keyed so the chat (and the typed text) survives the keyboard
      // changing the children around it.
      Expanded(
        key: const ValueKey('chat'),
        flex: 2,
        child: ChatPanel(roomId: roomId, compact: keyboardOpen),
      ),
    ],
  );
}
