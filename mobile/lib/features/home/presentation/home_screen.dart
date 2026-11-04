/// The Home tab: web's home page plus its left-bar channel lists as sections.
///
/// Sections, each loading and failing on its own ([FeedSection]): live
/// streams, followed channels, recommended channels (those not already
/// followed, like web) and popular videos. Pull-to-refresh reloads all of
/// them, and so does coming back from the background. Cards are not tappable
/// until the profile and player routes exist.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/home/home_providers.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// Home feed screen.
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  late final AppLifecycleListener _lifecycle;

  @override
  void initState() {
    super.initState();
    // Anything may have changed while the app was in the background.
    _lifecycle = AppLifecycleListener(onResume: () => refreshHome(ref));
  }

  @override
  void dispose() {
    _lifecycle.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final live = ref.watch(popularLiveProvider);
    final vods = ref.watch(popularVodsProvider);
    final following = ref.watch(followingChannelsProvider);
    final recommended = ref.watch(recommendedChannelsProvider);
    final now = DateTime.now();

    // Web hides followed channels from the recommended list.
    final followedIds = {
      for (final user in following.value ?? const <PublicUser>[]) user.id,
    };
    final recommendedNotFollowed = recommended.whenData(
      (users) => [
        for (final user in users)
          if (!followedIds.contains(user.id)) user,
      ],
    );

    return Scaffold(
      appBar: AppBar(title: Text(l10n.common_app_title)),
      body: RefreshIndicator(
        onRefresh: () => refreshHome(ref),
        child: ListView(
          // Lets pull-to-refresh work when the content is short.
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            FeedSection(
              key: const ValueKey('section-live'),
              title: l10n.common_livestreams,
              value: live,
              onRetry: () => ref.invalidate(popularLiveProvider),
              empty: EmptyNotice(
                icon: Icons.play_circle_outline,
                title: l10n.common_no_livestreams,
                description: l10n.common_no_livestreams_description,
              ),
              builder: (context, streams) => _CardList(
                cards: [
                  for (final stream in streams)
                    MediaCard(
                      key: ValueKey('live-${stream.id}'),
                      thumbnailUrl: mediaThumbnailUrl(
                        apiBaseUrl,
                        stream.id,
                        stream.thumbnailUrl,
                      ),
                      title: stream.title,
                      userId: stream.userId,
                      meta:
                          '${viewsLabel(l10n, stream.viewCount)} · '
                          '${l10n.common_started_at(timeAgo(l10n, stream.startedAt, now))}',
                    ),
                ],
              ),
            ),
            FeedSection(
              key: const ValueKey('section-following'),
              title: l10n.common_following,
              value: following,
              onRetry: () => ref.invalidate(followingChannelsProvider),
              builder: (context, users) => _ChannelRow(users: users),
            ),
            FeedSection(
              key: const ValueKey('section-recommended'),
              title: l10n.common_recommended,
              value: recommendedNotFollowed,
              onRetry: () => ref.invalidate(recommendedChannelsProvider),
              builder: (context, users) => _ChannelRow(users: users),
            ),
            FeedSection(
              key: const ValueKey('section-vods'),
              title: l10n.common_videos,
              value: vods,
              onRetry: () => ref.invalidate(popularVodsProvider),
              empty: EmptyNotice(
                icon: Icons.movie_outlined,
                title: l10n.common_no_videos,
                description: l10n.common_no_videos_description,
              ),
              builder: (context, items) => _CardList(
                cards: [
                  for (final vod in items)
                    MediaCard(
                      key: ValueKey('vod-${vod.id}'),
                      thumbnailUrl: mediaThumbnailUrl(
                        apiBaseUrl,
                        vod.id,
                        vod.thumbnailUrl,
                      ),
                      title: vod.title,
                      userId: vod.userId,
                      durationLabel: formatDuration(vod.duration),
                      meta:
                          '${viewsLabel(l10n, vod.viewCount)} · '
                          '${timeAgo(l10n, vod.createdAt, now)}',
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

class _CardList extends StatelessWidget {
  const _CardList({required this.cards});

  final List<Widget> cards;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(horizontal: 16),
    child: Column(children: cards),
  );
}

/// Horizontal avatar + name strip of channels.
class _ChannelRow extends StatelessWidget {
  const _ChannelRow({required this.users});

  final List<PublicUser> users;

  @override
  Widget build(BuildContext context) => SizedBox(
    height: 88,
    child: ListView.separated(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: users.length,
      separatorBuilder: (_, _) => const SizedBox(width: 16),
      itemBuilder: (context, index) {
        final user = users[index];
        return SizedBox(
          width: 64,
          child: Column(
            children: [
              UserAvatar(
                pictureUrl: user.profilePicture,
                name: user.username,
                radius: 28,
              ),
              const SizedBox(height: 4),
              Text(
                user.username,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        );
      },
    ),
  );
}
