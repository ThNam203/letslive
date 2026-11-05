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
    this.total,
  });

  /// How many items the server says the whole list has (`meta.total` of the
  /// first page); null when the endpoint reports none. Notifiers that add or
  /// remove items keep it in step.
  final int? total;

  /// The items loaded so far, in order, without duplicates.
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
    int? total,
    bool clearError = false,
  }) => PagedFeed(
    total: total ?? this.total,
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
  ///
  /// Implementations use `ref.read`: the repository provider never changes,
  /// and this also runs outside `build` (from [loadMore]), where `watch` is
  /// not allowed.
  Future<ApiResponse<List<T>>> fetchPage(int page);

  @override
  Future<PagedFeed<T>> build() async {
    final response = await fetchPage(0);
    final first = unwrapList(response);
    return PagedFeed(
      items: first,
      hasMore: first.length >= pageSize,
      total: response.meta?.total,
    );
  }

  /// Fetches page 0 again and merges it into what is loaded: the fresh
  /// items come first, in server order, and the items of the later pages
  /// already loaded stay. The loaded pages and the scroll position survive,
  /// unlike a rebuild. A failed fetch changes nothing. For lists kept current
  /// from realtime events (catch-up after a reconnect).
  Future<void> refreshFirstPage() async {
    final response = await fetchPage(0);
    final fresh = response.data;
    final feed = state.value;
    if (!ref.mounted || !response.success || fresh == null || feed == null) {
      return;
    }
    final freshIds = {for (final item in fresh) idOf(item)};
    state = AsyncData(
      feed.copyWith(
        items: [
          ...fresh,
          for (final item in feed.items)
            if (!freshIds.contains(idOf(item))) item,
        ],
        total: response.meta?.total,
      ),
    );
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
