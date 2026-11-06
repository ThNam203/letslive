/// The Home tab: web's home page.
///
/// Sections, each loading and failing on its own: live streams and popular
/// videos. Web's left-bar channel lists (followed and recommended) live in
/// the People tab of the user's profile (`profile_people_tab.dart`). The videos load page by page as the user nears the end of
/// the list. Pull-to-refresh reloads everything, and so does coming back after
/// a real stretch in the background. Live cards and channels open the channel
/// (`/users/:userId`); video cards open the video
/// (`/users/:userId/vods/:vodId`).
library;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/router/app_chrome.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/router/header_actions.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/home/home_providers.dart';
import 'package:letslive_mobile/features/home/presentation/media_card.dart';

/// How close to the end of the list (in logical pixels) the next page of
/// videos starts loading: the list builds the [PagedFooter], which triggers
/// the load, this far ahead of the viewport.
const double _loadMoreDistance = 300;

/// Horizontal screen padding of the feed cards.
const EdgeInsets _cardPadding = EdgeInsets.symmetric(horizontal: 16);

/// Home feed screen.
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  late final AppLifecycleListener _lifecycle;

  /// When the app last left the foreground; null while it is in front.
  DateTime? _hiddenAt;

  @override
  void initState() {
    super.initState();
    // `hide` means really in the background; `inactive` blips (control
    // center, permission dialogs) never reach it.
    _lifecycle = AppLifecycleListener(
      onHide: () => _hiddenAt = ref.read(homeClockProvider)(),
      onResume: _refreshAfterBackground,
    );
  }

  void _refreshAfterBackground() {
    final hiddenAt = _hiddenAt;
    _hiddenAt = null;
    if (hiddenAt == null) return;
    final away = ref.read(homeClockProvider)().difference(hiddenAt);
    if (away >= homeRefreshAfterBackground) {
      refreshHome(ref);
    }
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
    final now = DateTime.now();

    // The rows are plain widgets in a lazy list: only the visible cards (plus
    // a little ahead) are built, however many pages of videos were loaded.
    final rows = <Widget>[
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
        builder: (context, streams) => Padding(
          padding: _cardPadding,
          child: Column(
            children: [
              for (final stream in streams)
                _liveCard(l10n, apiBaseUrl, stream, now),
            ],
          ),
        ),
      ),
      SectionHeader(l10n.common_videos),
      ..._vodRows(l10n, apiBaseUrl, vods, now),
    ];

    return TabScaffold(
      appBar: AppBar(
        title: Text(l10n.common_app_title),
        actions: const [HeaderActions()],
      ),
      body: RefreshIndicator(
        onRefresh: () => refreshHome(ref),
        child: ListView.builder(
          // Lets pull-to-refresh work when the content is short.
          physics: const AlwaysScrollableScrollPhysics(),
          scrollCacheExtent: const ScrollCacheExtent.pixels(_loadMoreDistance),
          itemCount: rows.length,
          itemBuilder: (context, index) => rows[index],
        ),
      ),
    );
  }

  Widget _liveCard(
    AppLocalizations l10n,
    String apiBaseUrl,
    Livestream stream,
    DateTime now,
  ) => MediaCard(
    key: ValueKey('live-${stream.id}'),
    thumbnailUrl: mediaThumbnailUrl(apiBaseUrl, stream.id, stream.thumbnailUrl),
    title: stream.title,
    userId: stream.userId,
    meta:
        '${viewsLabel(l10n, stream.viewCount)} · '
        '${l10n.common_started_at(timeAgo(l10n, stream.startedAt, now))}',
    onTap: () => context.push(AppRoutes.channel(stream.userId)),
  );

  /// The body of the videos section: status (loading, error, empty) or one
  /// row per video followed by the paging footer.
  List<Widget> _vodRows(
    AppLocalizations l10n,
    String apiBaseUrl,
    AsyncValue<PagedFeed<Vod>> vods,
    DateTime now,
  ) {
    final status = feedStatusBody(
      value: vods.whenData((feed) => feed.items),
      onRetry: () => ref.invalidate(popularVodsProvider),
      empty: EmptyNotice(
        icon: Icons.movie_outlined,
        title: l10n.common_no_videos,
        description: l10n.common_no_videos_description,
      ),
    );
    if (status != null) return [status];
    return [
      for (final vod in vods.requireValue.items)
        Padding(
          padding: _cardPadding,
          child: _vodCard(l10n, apiBaseUrl, vod, now),
        ),
      PagedFooter(
        feed: vods.value,
        onLoadMore: ref.read(popularVodsProvider.notifier).loadMore,
      ),
    ];
  }

  Widget _vodCard(
    AppLocalizations l10n,
    String apiBaseUrl,
    Vod vod,
    DateTime now,
  ) => MediaCard(
    key: ValueKey('vod-${vod.id}'),
    thumbnailUrl: mediaThumbnailUrl(apiBaseUrl, vod.id, vod.thumbnailUrl),
    title: vod.title,
    userId: vod.userId,
    durationLabel: formatDuration(vod.duration),
    meta:
        '${viewsLabel(l10n, vod.viewCount)} · '
        '${timeAgo(l10n, vod.createdAt, now)}',
    onTap: () => context.push(AppRoutes.vod(vod.userId, vod.id)),
  );
}
