/// Riverpod providers for the home feed. Each section loads on its own, so
/// one failing request only shows an error in that section.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/home/data/home_repository.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// Live streams and popular videos.
final homeRepositoryProvider = Provider<HomeRepository>(
  (ref) => HomeRepository(ref.watch(apiClientProvider)),
);

/// Streams that are live now.
final popularLiveProvider = FutureProvider<List<Livestream>>(
  (ref) async =>
      unwrapList(await ref.watch(homeRepositoryProvider).popularLive()),
  retry: noRetry,
);

/// Popular videos.
final popularVodsProvider = FutureProvider<List<Vod>>(
  (ref) async =>
      unwrapList(await ref.watch(homeRepositoryProvider).popularVods()),
  retry: noRetry,
);

/// Reloads every home section and the owners shown on the cards. Completes
/// when the lists settled; failures are not thrown here, each section shows
/// its own error.
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
