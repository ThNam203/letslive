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
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
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
          onStopped: ref.read(channelProvider(userId).notifier).recheck,
        ),
        ChannelOffline(:final user) => _NoticeWithChat(
          roomId: user.id,
          notice: _Notice(
            icon: Icons.videocam_off_outlined,
            title: l10n.users_offline,
            message: l10n.users_channel_offline_notice(user.username),
          ),
        ),
        ChannelEnded(:final user) => _NoticeWithChat(
          roomId: user.id,
          notice: _Notice(
            icon: Icons.live_tv_outlined,
            message: l10n.users_channel_stream_ended,
          ),
        ),
        ChannelNotFound() => _Notice(
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

/// The player, then the stream title and the streamer's name.
class _LiveBody extends ConsumerWidget {
  const _LiveBody({
    required this.user,
    required this.stream,
    required this.onStopped,
  });

  final PublicUser user;
  final Livestream stream;
  final Future<void> Function() onStopped;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    return Column(
      children: [
        LivePlayerView(
          // A new stream (the user restarted) gets a new player.
          key: ValueKey(stream.id),
          url: liveStreamUrl(apiBaseUrl, stream.id),
          title: stream.title,
          author: user.username,
          onStopped: onStopped,
        ),
        Padding(
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
        ),
        const Divider(height: 1),
        // The room id is the streamer's user id.
        Expanded(child: ChatPanel(roomId: user.id)),
      ],
    );
  }
}

/// The offline / ended [notice] above the room's chat, which web also shows
/// when nobody is streaming.
class _NoticeWithChat extends StatelessWidget {
  const _NoticeWithChat({required this.roomId, required this.notice});

  final String roomId;
  final Widget notice;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      Expanded(child: notice),
      const Divider(height: 1),
      Expanded(flex: 2, child: ChatPanel(roomId: roomId)),
    ],
  );
}

/// A centered icon, optional [title], [message] and a back button. Back
/// pops to where the user came from, or goes home after a deep link.
class _Notice extends StatelessWidget {
  const _Notice({required this.icon, required this.message, this.title});

  final IconData icon;
  final String? title;
  final String message;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final heading = title;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Icon(icon, size: 48, color: theme.colorScheme.onSurfaceVariant),
        const SizedBox(height: 12),
        if (heading != null)
          Text(
            heading,
            textAlign: TextAlign.center,
            style: theme.textTheme.titleMedium,
          ),
        Text(message, textAlign: TextAlign.center),
        const SizedBox(height: 16),
        Center(
          child: FilledButton.tonal(
            onPressed: () =>
                context.canPop() ? context.pop() : context.go(AppRoutes.home),
            child: Text(l10n.users_channel_back),
          ),
        ),
      ],
    );
  }
}
