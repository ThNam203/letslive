/// The channel screen, `/users/:userId` (web `app/(main)/users/[userId]`):
/// the live player and title while the user streams, otherwise an offline
/// notice. A stream that ends while watched turns into an "ended" notice;
/// an unknown user shows "not found". Every notice has a way back. Held
/// sideways, a live channel shows only the player (the fullscreen mode).
///
/// Under the player (or the offline / ended notice) sits the profile
/// (`features/profile`): header with the follow button and the tabs Chat /
/// Home / Videos / About, Chat being the default one. How the end of a
/// stream is noticed is described in `watch_providers.dart`.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/widgets/back_notice.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_header.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_tabs.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
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
          onRetry: () {
            ref.invalidate(publicUserProvider(userId));
            ref.invalidate(channelProvider(userId));
          },
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
          user: user,
          keyboardOpen: keyboardOpen,
          notice: BackNotice(
            icon: Icons.videocam_off_outlined,
            title: l10n.users_offline,
            message: l10n.users_channel_offline_notice(user.username),
          ),
        ),
        ChannelEnded(:final user) => _NoticeWithChat(
          user: user,
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

    // Sideways, a live channel shows only its player (see _LiveBody).
    final playerOnly = isLandscape(context) && body is _LiveBody;
    return Scaffold(
      appBar: playerOnly
          ? null
          : AppBar(
              title: user == null
                  ? null
                  : Text(
                      user.username,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
            ),
      body: body,
    );
  }
}

/// Tab content height below which the live layout scrolls instead of
/// squeezing it: room for a few chat lines plus the message box (and its
/// suggestions).
const double _minChatHeight = 240;
const double _minCompactChatHeight = 200;

/// The stream title block's vertical padding (16 + 16) plus 1 px of slack
/// for rounding.
const double _infoPadding = 33;

/// `titleMedium`'s font size and line height (Material 3), and the title's
/// line limit.
const double _titleFontSize = 16;
const double _titleLineHeight = 1.5;
const int _titleMaxLines = 2;

/// The player and the stream title, then the profile tabs with the chat.
///
/// With the keyboard open ([keyboardOpen]) the title block, the profile
/// header and the tab bar go and only the chat remains. If the tabs would
/// still get less than their minimum height the page scrolls (player on
/// top, tabs at the minimum) rather than overflow. Held sideways only the
/// player shows: rotating is the fullscreen mode. Player and chat hold
/// [GlobalKey]s so switching between these layouts keeps playback and the
/// typed text.
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

  /// An upper bound of the title block's height at the current text scale:
  /// its padding plus [_titleMaxLines] lines of `titleMedium`.
  double _infoHeight(BuildContext context) {
    final scaler = MediaQuery.textScalerOf(context);
    return _infoPadding +
        scaler.scale(_titleFontSize) * _titleLineHeight * _titleMaxLines;
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
    // Rotating is the fullscreen mode: sideways only the player shows.
    if (isLandscape(context)) return LandscapePlayer(player: player);
    final info = Padding(
      padding: const EdgeInsets.all(16),
      child: Text(
        stream.title,
        maxLines: _titleMaxLines,
        overflow: TextOverflow.ellipsis,
        style: theme.textTheme.titleMedium,
      ),
    );
    // The room id is the streamer's user id.
    final chat = ChatPanel(key: _chatKey, roomId: user.id, compact: true);
    final tabs = ProfileTabs(
      user: user,
      chat: chat,
      showBanner: false,
      compact: keyboard,
    );
    final minTabs = keyboard
        ? _minCompactChatHeight
        : _minChatHeight + _profileChromeHeight(context);

    return LayoutBuilder(
      builder: (context, box) {
        final used =
            box.maxWidth * 9 / 16 + (keyboard ? 0 : _infoHeight(context));
        if (box.maxHeight - used >= minTabs) {
          return Column(
            children: [
              player,
              if (!keyboard) info,
              Expanded(child: tabs),
            ],
          );
        }
        return SingleChildScrollView(
          child: Column(
            children: [
              player,
              if (!keyboard) info,
              SizedBox(height: minTabs, child: tabs),
            ],
          ),
        );
      },
    );
  }
}

/// The offline / ended [notice] above the room's chat, which web also shows
/// when nobody is streaming, inside the profile (banner and tabs). With the
/// keyboard open only the chat remains.
class _NoticeWithChat extends StatelessWidget {
  const _NoticeWithChat({
    required this.user,
    required this.notice,
    required this.keyboardOpen,
  });

  final PublicUser user;
  final Widget notice;
  final bool keyboardOpen;

  @override
  Widget build(BuildContext context) => ProfileTabs(
    user: user,
    compact: keyboardOpen,
    chat: Column(
      children: [
        if (!keyboardOpen) ...[
          Expanded(child: notice),
          const Divider(height: 1),
        ],
        // Keyed so the chat (and the typed text) survives the keyboard
        // changing the children around it.
        Expanded(
          key: const ValueKey('chat'),
          flex: 2,
          child: ChatPanel(roomId: user.id, compact: true),
        ),
      ],
    ),
  );
}

/// An upper bound of the profile header and tab bar height at the current
/// text scale.
double _profileChromeHeight(BuildContext context) =>
    profileHeaderHeight(context) + profileTabBarHeight;
