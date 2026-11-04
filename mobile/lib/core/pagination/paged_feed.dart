/// Infinite-scroll state and notifier shared by every paged list (popular
/// videos on Home, the type tabs of Search).
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';

/// The items of a paged list loaded so far.
class PagedFeed<T> {
  const PagedFeed({
    required this.items,
    required this.hasMore,
    this.nextPage = 1,
    this.isLoadingMore = false,
    this.loadMoreError,
  });

  final List<T> items;

  /// False once a page came back shorter than a full page.
  final bool hasMore;

  /// Page index the next [PagedFeedNotifier.loadMore] requests.
  final int nextPage;
  final bool isLoadingMore;

  /// The last page request failed; the footer offers a retry.
  final ApiLoadException? loadMoreError;

  PagedFeed<T> copyWith({
    List<T>? items,
    bool? hasMore,
    int? nextPage,
    bool? isLoadingMore,
    ApiLoadException? loadMoreError,
    bool clearError = false,
  }) => PagedFeed(
    items: items ?? this.items,
    hasMore: hasMore ?? this.hasMore,
    nextPage: nextPage ?? this.nextPage,
    isLoadingMore: isLoadingMore ?? this.isLoadingMore,
    loadMoreError: clearError ? null : loadMoreError ?? this.loadMoreError,
  );
}

/// Loads page 0, then more with [loadMore]. A failed first page is the
/// provider's error; a failed later page keeps the list and sets
/// [PagedFeed.loadMoreError]. Subclasses say how to fetch a page.
abstract class PagedFeedNotifier<T> extends AsyncNotifier<PagedFeed<T>> {
  /// Items per full page; a shorter page is the last one.
  int get pageSize;

  /// Identity of an item, to drop duplicates when new items shift the pages.
  String idOf(T item);

  /// Requests page [page] (0-based).
  Future<ApiResponse<List<T>>> fetchPage(int page);

  @override
  Future<PagedFeed<T>> build() async {
    final first = unwrapList(await fetchPage(0));
    return PagedFeed(items: first, hasMore: first.length >= pageSize);
  }

  /// Appends the next page. Does nothing while a page is loading, after the
  /// last page, or before the first page arrived. Also the retry after an
  /// error.
  Future<void> loadMore() async {
    if (!ref.mounted) return;
    final feed = state.value;
    if (feed == null || !feed.hasMore || feed.isLoadingMore) return;
    final loading = feed.copyWith(isLoadingMore: true, clearError: true);
    state = AsyncData(loading);
    final response = await fetchPage(feed.nextPage);
    // Refreshed or disposed meanwhile: this page belongs to an old feed.
    if (!ref.mounted || !identical(state.value, loading)) return;
    if (!response.success) {
      state = AsyncData(
        loading.copyWith(
          isLoadingMore: false,
          loadMoreError: ApiLoadException(response),
        ),
      );
      return;
    }
    final page = response.data ?? <T>[];
    final known = {for (final item in feed.items) idOf(item)};
    state = AsyncData(
      loading.copyWith(
        items: [
          ...feed.items,
          for (final item in page)
            if (!known.contains(idOf(item))) item,
        ],
        hasMore: page.length >= pageSize,
        nextPage: feed.nextPage + 1,
        isLoadingMore: false,
      ),
    );
  }
}
