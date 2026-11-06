/// Riverpod providers for the home feed. Each section loads on its own, so
/// one failing request only shows an error in that section.
///
/// The data providers are `autoDispose`: leaving the signed-in shell (logout)
/// drops them, so the next account never sees the previous one's data.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
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

/// Loads page 0 of the popular videos, then more with `loadMore`.
class PopularVodsNotifier extends PagedFeedNotifier<Vod> {
  @override
  int get pageSize => popularVodsPageSize;

  @override
  String idOf(Vod item) => item.id;

  @override
  Future<ApiResponse<List<Vod>>> fetchPage(int page) =>
      ref.read(homeRepositoryProvider).popularVods(page: page);
}

/// Popular videos, paginated.
final popularVodsProvider =
    AsyncNotifierProvider.autoDispose<PopularVodsNotifier, PagedFeed<Vod>>(
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
  ];
  return Future.wait([
    for (final reload in reloads)
      reload.then<void>((_) {}, onError: (Object _) {}),
  ]);
}
