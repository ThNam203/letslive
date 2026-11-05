import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// Videos listed under "More videos" (web `MORE_VIDEOS_COUNT`).
const int _moreVideosCount = 4;

/// The Home tab of a profile (web `ProfileHomeTab`): the latest video
/// featured, then up to four more with a "View all" that calls [onViewAll].
/// Loading, error with retry and the empty state come from
/// [feedStatusBody].
class ProfileHomeTab extends ConsumerWidget {
  const ProfileHomeTab({
    required this.user,
    required this.onViewAll,
    super.key,
  });

  /// The profile owner.
  final PublicUser user;

  /// Switches to the Videos tab.
  final VoidCallback onViewAll;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final vods = ref
        .watch(channelVodsProvider(user.id))
        .whenData((feed) => feed.items);
    final status = feedStatusBody(
      value: vods,
      onRetry: () => ref.invalidate(channelVodsProvider(user.id)),
      empty: EmptyNotice(
        icon: Icons.movie_outlined,
        title: l10n.users_channel_no_videos,
        description: l10n.users_channel_no_videos_description(user.username),
      ),
    );
    if (status != null) return SingleChildScrollView(child: status);

    final [latest, ...rest] = vods.requireValue;
    final more = rest.take(_moreVideosCount);
    final now = DateTime.now();
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          l10n.users_channel_latest_video,
          style: theme.textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        _FeaturedVod(vod: latest, apiBaseUrl: apiBaseUrl, now: now),
        if (more.isNotEmpty) ...[
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: Text(
                  l10n.users_channel_more_videos,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.titleMedium,
                ),
              ),
              TextButton(
                key: const ValueKey('view-all-videos'),
                onPressed: onViewAll,
                child: Text(l10n.users_channel_view_all),
              ),
            ],
          ),
          for (final vod in more)
            channelVodCard(context, l10n, apiBaseUrl, vod),
        ],
      ],
    );
  }
}

/// A video card of a channel: opens the video.
Widget channelVodCard(
  BuildContext context,
  AppLocalizations l10n,
  String apiBaseUrl,
  Vod vod,
) => MediaCard(
  key: ValueKey('vod-${vod.id}'),
  thumbnailUrl: mediaThumbnailUrl(apiBaseUrl, vod.id, vod.thumbnailUrl),
  title: vod.title,
  userId: vod.userId,
  durationLabel: formatDuration(vod.duration),
  meta:
      '${viewsLabel(l10n, vod.viewCount)} · '
      '${timeAgo(l10n, vod.createdAt, DateTime.now())}',
  onTap: () => context.push(AppRoutes.vod(vod.userId, vod.id)),
);

/// The latest video: big thumbnail, title, meta, description and a Watch
/// button (web `FeaturedVod`).
class _FeaturedVod extends StatelessWidget {
  const _FeaturedVod({
    required this.vod,
    required this.apiBaseUrl,
    required this.now,
  });

  final Vod vod;
  final String apiBaseUrl;
  final DateTime now;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final description = vod.description?.trim() ?? '';
    void open() => context.push(AppRoutes.vod(vod.userId, vod.id));
    return Column(
      key: const ValueKey('featured-vod'),
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        InkWell(
          onTap: open,
          child: MediaThumbnail(
            url: mediaThumbnailUrl(apiBaseUrl, vod.id, vod.thumbnailUrl),
            durationLabel: formatDuration(vod.duration),
          ),
        ),
        const SizedBox(height: 8),
        Text(
          vod.title,
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
          style: theme.textTheme.titleLarge,
        ),
        Text(
          '${viewsLabel(l10n, vod.viewCount)} · '
          '${timeAgo(l10n, vod.createdAt, now)}',
          style: theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
        if (description.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Text(
              description,
              maxLines: 3,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        const SizedBox(height: 8),
        FilledButton.icon(
          onPressed: open,
          icon: const Icon(Icons.play_arrow),
          label: Text(l10n.users_channel_watch),
        ),
      ],
    );
  }
}
