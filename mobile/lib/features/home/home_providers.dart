/// Riverpod providers for the home feed. Each section loads on its own, so
/// one failing request only shows an error in that section.
///
/// The data providers are `autoDispose`: leaving the signed-in shell (logout)
/// drops them, so the next account never sees the previous one's data.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/home/data/home_repository.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// How long the app must have been in the background before coming back
/// reloads the home feed (shorter blips, like the control center, do not).
const Duration homeRefreshAfterBackground = Duration(seconds: 30);

/// The current time; tests override it to skip the wait in the background.
final homeClockProvider = Provider<DateTime Function()>((ref) => DateTime.now);

/// Live streams and popular videos.
final homeRepositoryProvider = Provider<HomeRepository>(
  (ref) => HomeRepository(ref.watch(apiClientProvider)),
);

/// Streams that are live now.
final popularLiveProvider = FutureProvider.autoDispose<List<Livestream>>(
  (ref) async =>
      unwrapList(await ref.watch(homeRepositoryProvider).popularLive()),
  retry: noRetry,
);

/// The popular videos loaded so far (infinite scroll, like web).
class VodFeed {
  const VodFeed({
    required this.items,
    required this.hasMore,
    this.nextPage = 1,
    this.isLoadingMore = false,
    this.loadMoreError,
  });

  final List<Vod> items;

  /// False once a page came back shorter than a full page.
  final bool hasMore;

  /// Page index the next [PopularVodsNotifier.loadMore] requests.
  final int nextPage;
  final bool isLoadingMore;

  /// The last page request failed; the footer offers a retry.
  final ApiLoadException? loadMoreError;

  VodFeed copyWith({
    List<Vod>? items,
    bool? hasMore,
    int? nextPage,
    bool? isLoadingMore,
    ApiLoadException? loadMoreError,
    bool clearError = false,
  }) => VodFeed(
    items: items ?? this.items,
    hasMore: hasMore ?? this.hasMore,
    nextPage: nextPage ?? this.nextPage,
    isLoadingMore: isLoadingMore ?? this.isLoadingMore,
    loadMoreError: clearError ? null : loadMoreError ?? this.loadMoreError,
  );
}

/// Loads page 0 of the popular videos, then more with [loadMore]. A failed
/// first page is the provider's error; a failed later page keeps the list and
/// sets [VodFeed.loadMoreError].
class PopularVodsNotifier extends AsyncNotifier<VodFeed> {
  @override
  Future<VodFeed> build() async {
    final first = unwrapList(
      await ref.watch(homeRepositoryProvider).popularVods(),
    );
    return VodFeed(items: first, hasMore: first.length >= popularVodsPageSize);
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
    final response = await ref
        .read(homeRepositoryProvider)
        .popularVods(page: feed.nextPage);
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
    final page = response.data ?? const <Vod>[];
    // New videos shift the pages, so a video can show up twice.
    final known = {for (final vod in feed.items) vod.id};
    state = AsyncData(
      loading.copyWith(
        items: [
          ...feed.items,
          for (final vod in page)
            if (!known.contains(vod.id)) vod,
        ],
        hasMore: page.length >= popularVodsPageSize,
        nextPage: feed.nextPage + 1,
        isLoadingMore: false,
      ),
    );
  }
}

/// Popular videos, paginated.
final popularVodsProvider =
    AsyncNotifierProvider.autoDispose<PopularVodsNotifier, VodFeed>(
      PopularVodsNotifier.new,
      retry: noRetry,
    );

/// Reloads every home section and the owners shown on the cards; videos go
/// back to page 0. Completes when the lists settled; failures are not thrown
/// here, each section shows its own error.
Future<void> refreshHome(WidgetRef ref) {
  ref.invalidate(publicUserProvider);
  final reloads = [
    ref.refresh(popularLiveProvider.future),
    ref.refresh(popularVodsProvider.future),
    ref.refresh(followingChannelsProvider.future),
    ref.refresh(recommendedChannelsProvider.future),
  ];
  return Future.wait([
    for (final reload in reloads)
      reload.then<void>((_) {}, onError: (Object _) {}),
  ]);
}
