/// The rows of the search results (web `search-result-rows.tsx`): a channel,
/// a live stream and a video, each with the matched text highlighted.
/// Channels and live streams open the channel (`/users/:userId`); videos are
/// not tappable until the VOD screen exists.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';
import 'package:letslive_mobile/features/search/domain/highlight.dart';
import 'package:letslive_mobile/features/search/presentation/highlighted_text.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// Width of the thumbnail in a video or live row.
const double _thumbnailWidth = 120;

/// A channel: avatar, username, follower count and bio. Opens the channel.
class UserResultRow extends StatelessWidget {
  const UserResultRow({required this.user, required this.query, super.key});

  /// The channel to show.
  final PublicUser user;

  /// The searched text, highlighted in the username.
  final String query;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final bio = user.bio;
    final muted = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    return InkWell(
      onTap: () => context.push(AppRoutes.channel(user.id)),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            UserAvatar(
              pictureUrl: user.profilePicture,
              name: user.username,
              radius: 28,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  HighlightedText(
                    text: user.username,
                    query: query,
                    style: theme.textTheme.titleMedium,
                  ),
                  Text(
                    l10n.common_followers_count(
                      compactCount(l10n, user.followerCount),
                      user.followerCount,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: muted,
                  ),
                  if (bio != null && bio.isNotEmpty)
                    Text(
                      bio,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: muted,
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// A live stream that matched the search. Opens the streamer's channel.
class LiveResultRow extends StatelessWidget {
  const LiveResultRow({required this.stream, required this.query, super.key});

  /// The live stream to show.
  final Livestream stream;

  /// The searched text, highlighted in the texts.
  final String query;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return _MediaResultRow(
      thumbnailId: stream.id,
      thumbnailPath: stream.thumbnailUrl,
      liveLabel: l10n.users_live,
      title: stream.title,
      description: stream.description,
      userId: stream.userId,
      query: query,
      onTap: () => context.push(AppRoutes.channel(stream.userId)),
      meta:
          '${l10n.common_search_viewers(compactCount(l10n, stream.viewCount), stream.viewCount)}'
          ' · ${l10n.common_started_at(timeAgo(l10n, stream.startedAt, DateTime.now()))}',
    );
  }
}

/// A video that matched the search. Opens the video.
class VodResultRow extends StatelessWidget {
  const VodResultRow({required this.vod, required this.query, super.key});

  /// The video to show.
  final Vod vod;

  /// The searched text, highlighted in the texts.
  final String query;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return _MediaResultRow(
      thumbnailId: vod.id,
      thumbnailPath: vod.thumbnailUrl,
      durationLabel: formatDuration(vod.duration),
      title: vod.title,
      description: vod.description,
      userId: vod.userId,
      query: query,
      onTap: () => context.push(AppRoutes.vod(vod.userId, vod.id)),
      meta:
          '${viewsLabel(l10n, vod.viewCount)}'
          ' · ${timeAgo(l10n, vod.createdAt, DateTime.now())}',
    );
  }
}

/// Thumbnail on the left; title, streamer, meta line and a description
/// excerpt around the match on the right. The streamer is looked up by
/// [userId] ([publicUserProvider]); until it arrives the name is left out.
class _MediaResultRow extends ConsumerWidget {
  const _MediaResultRow({
    required this.thumbnailId,
    required this.thumbnailPath,
    required this.title,
    required this.description,
    required this.userId,
    required this.query,
    required this.meta,
    this.durationLabel,
    this.liveLabel,
    this.onTap,
  });

  final String thumbnailId;
  final String? thumbnailPath;
  final String title;
  final String? description;
  final String userId;
  final String query;
  final String meta;
  final String? durationLabel;
  final String? liveLabel;

  /// Called when the row is tapped; null leaves it inert.
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final streamer = ref.watch(publicUserProvider(userId)).value;
    final text = description;
    final snippet = text == null ? null : excerptAroundMatch(text, query);
    final muted = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    final row = Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: _thumbnailWidth,
            child: MediaThumbnail(
              url: mediaThumbnailUrl(apiBaseUrl, thumbnailId, thumbnailPath),
              durationLabel: durationLabel,
              liveLabel: liveLabel,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                HighlightedText(
                  text: title,
                  query: query,
                  maxLines: 2,
                  style: theme.textTheme.titleSmall,
                ),
                if (streamer != null)
                  HighlightedText(
                    text: streamer.username,
                    query: query,
                    style: muted,
                  ),
                Text(
                  meta,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: muted,
                ),
                if (snippet != null)
                  HighlightedText(
                    text: snippet,
                    query: query,
                    maxLines: 2,
                    style: muted,
                  ),
              ],
            ),
          ),
        ],
      ),
    );
    return onTap == null ? row : InkWell(onTap: onTap, child: row);
  }
}
