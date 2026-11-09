import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/profile/presentation/profile_home_tab.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// How close to the end of the list (logical pixels) the next page starts
/// loading (see `PagedFooter`).
const double _loadMoreDistance = 300;

/// The Videos tab of a profile (web `ProfileVideosTab`): every public video
/// of the user, loaded page by page as the list nears its end. A failed
/// first page shows an error with retry, a failed later page an inline retry
/// under the loaded videos.
class ProfileVideosTab extends ConsumerWidget {
  const ProfileVideosTab({required this.user, super.key});

  /// The profile owner.
  final PublicUser user;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final feed = ref.watch(channelVodsProvider(user.id));
    final status = feedStatusBody(
      value: feed.whenData((f) => f.items),
      onRetry: () => ref.invalidate(channelVodsProvider(user.id)),
      empty: EmptyNotice(
        icon: Icons.movie_outlined,
        title: l10n.users_channel_no_videos,
        description: l10n.users_channel_no_videos_description(user.username),
      ),
    );
    if (status != null) return SingleChildScrollView(child: status);

    final items = feed.requireValue.items;
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      scrollCacheExtent: const ScrollCacheExtent.pixels(_loadMoreDistance),
      itemCount: items.length + 1,
      itemBuilder: (context, index) => index < items.length
          ? channelVodCard(context, l10n, apiBaseUrl, items[index])
          : PagedFooter(
              feed: feed.value,
              onLoadMore: ref
                  .read(channelVodsProvider(user.id).notifier)
                  .loadMore,
            ),
    );
  }
}
