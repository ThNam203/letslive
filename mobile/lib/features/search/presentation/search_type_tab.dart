import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/search/presentation/search_result_rows.dart';
import 'package:letslive_mobile/features/search/search_providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// How close to the end of the list (in logical pixels) the next page starts
/// loading: the list builds the [PagedFooter], which triggers the load, this
/// far ahead of the viewport.
const double _loadMoreDistance = 300;

/// Channels tab: infinite list of channels matching [query].
class UsersTab extends ConsumerWidget {
  const UsersTab({required this.query, super.key});

  final String query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = usersSearchProvider(query);
    return PagedResults<PublicUser>(
      value: ref.watch(provider),
      onRetry: () => ref.invalidate(provider),
      onLoadMore: ref.read(provider.notifier).loadMore,
      emptyTitle: AppLocalizations.of(context)
          .common_search_no_results_type_users,
      rowBuilder: (user) =>
          UserResultRow(key: ValueKey(user.id), user: user, query: query),
    );
  }
}

/// Live tab: infinite list of live streams matching [query].
class LivesTab extends ConsumerWidget {
  const LivesTab({required this.query, super.key});

  final String query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = livesSearchProvider(query);
    return PagedResults<Livestream>(
      value: ref.watch(provider),
      onRetry: () => ref.invalidate(provider),
      onLoadMore: ref.read(provider.notifier).loadMore,
      emptyTitle: AppLocalizations.of(context)
          .common_search_no_results_type_lives,
      rowBuilder: (stream) =>
          LiveResultRow(key: ValueKey(stream.id), stream: stream, query: query),
    );
  }
}

/// Videos tab: infinite list of videos matching [query].
class VodsTab extends ConsumerWidget {
  const VodsTab({required this.query, super.key});

  final String query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = vodsSearchProvider(query);
    return PagedResults<Vod>(
      value: ref.watch(provider),
      onRetry: () => ref.invalidate(provider),
      onLoadMore: ref.read(provider.notifier).loadMore,
      emptyTitle: AppLocalizations.of(context)
          .common_search_no_results_type_vods,
      rowBuilder: (vod) =>
          VodResultRow(key: ValueKey(vod.id), vod: vod, query: query),
    );
  }
}

/// A lazy list of one result type: spinner while the first page loads, an
/// error with a retry, an empty notice ([emptyTitle] plus the spelling hint),
/// or the rows followed by the paging footer (spinner, or an inline retry
/// after a failed page).
class PagedResults<T> extends StatelessWidget {
  const PagedResults({
    required this.value,
    required this.onRetry,
    required this.onLoadMore,
    required this.rowBuilder,
    required this.emptyTitle,
    super.key,
  });

  final AsyncValue<PagedFeed<T>> value;

  /// Reloads from the first page.
  final VoidCallback onRetry;
  final VoidCallback onLoadMore;
  final Widget Function(T item) rowBuilder;
  final String emptyTitle;

  @override
  Widget build(BuildContext context) {
    final status = feedStatusBody(
      value: value.whenData((feed) => feed.items),
      onRetry: onRetry,
      empty: EmptyNotice(
        icon: Icons.search_off,
        title: emptyTitle,
        description: AppLocalizations.of(context).common_search_no_results_hint,
      ),
    );
    if (status != null) return ListView(children: [status]);
    final items = value.requireValue.items;
    return ListView.builder(
      scrollCacheExtent: const ScrollCacheExtent.pixels(_loadMoreDistance),
      itemCount: items.length + 1,
      itemBuilder: (context, index) => index < items.length
          ? rowBuilder(items[index])
          : PagedFooter(feed: value.value, onLoadMore: onLoadMore),
    );
  }
}
