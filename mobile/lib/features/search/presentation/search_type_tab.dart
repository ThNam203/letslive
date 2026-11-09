import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/search/search_providers.dart';

/// How close to the end of the list (in logical pixels) the next page starts
/// loading: the list builds the [PagedFooter], which triggers the load, this
/// far ahead of the viewport.
const double _loadMoreDistance = 300;

/// A type tab: an infinite list of the results [provider] loads. Spinner
/// while the first page loads, an error with a retry, an empty notice
/// ([emptyTitle] plus the spelling hint), or the rows from [rowBuilder]
/// followed by the paging footer (spinner, or an inline retry after a failed
/// page).
class SearchTypeTab<T> extends ConsumerWidget {
  const SearchTypeTab({
    required this.provider,
    required this.emptyTitle,
    required this.rowBuilder,
    super.key,
  });

  /// The paged results of the current query.
  final AsyncNotifierProvider<SearchNotifier<T>, PagedFeed<T>> provider;

  /// Shown when there are no results of this type.
  final String emptyTitle;

  /// Builds the row of one result.
  final Widget Function(T item) rowBuilder;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final value = ref.watch(provider);
    final status = feedStatusBody(
      value: value.whenData((feed) => feed.items),
      onRetry: () => ref.invalidate(provider),
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
          : PagedFooter(
              feed: value.value,
              onLoadMore: ref.read(provider.notifier).loadMore,
            ),
    );
  }
}
