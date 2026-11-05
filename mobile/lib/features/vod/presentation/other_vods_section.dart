import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';

/// "Other streams": the uploader's other public videos (web's sidebar, which
/// moves under the video on small screens). Hidden when there are none.
class OtherVodsSection extends ConsumerWidget {
  const OtherVodsSection({
    required this.userId,
    required this.currentVodId,
    super.key,
  });

  /// The uploader, from the route.
  final String userId;

  /// The video being watched, left out of the list.
  final String currentVodId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final apiBaseUrl = ref.watch(appConfigProvider).apiBaseUrl;
    final vods = ref
        .watch(userVodsProvider(userId))
        .whenData(
          (all) => [
            for (final vod in all)
              if (vod.id != currentVodId) vod,
          ],
        );
    final now = DateTime.now();
    return FeedSection(
      title: l10n.common_other_streams,
      value: vods,
      onRetry: () => ref.invalidate(userVodsProvider(userId)),
      builder: (context, items) => Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Column(
          children: [
            for (final vod in items)
              MediaCard(
                key: ValueKey('other-${vod.id}'),
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
                onTap: () => context.push(AppRoutes.vod(vod.userId, vod.id)),
              ),
          ],
        ),
      ),
    );
  }
}
