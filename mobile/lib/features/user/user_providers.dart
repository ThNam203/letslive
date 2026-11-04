/// Riverpod providers for public user data.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/data/user_repository.dart';

/// Public user lookups.
final userRepositoryProvider = Provider<UserRepository>(
  (ref) => UserRepository(ref.watch(apiClientProvider)),
);

/// One user by id. Cards of the same owner share one request (like web's
/// per-id query cache); invalidate the family to reload.
final publicUserProvider = FutureProvider.family<PublicUser, String>(
  (ref, id) async =>
      unwrapData(await ref.watch(userRepositoryProvider).byId(id)),
  retry: noRetry,
);

/// Channels the signed-in user follows.
final followingChannelsProvider = FutureProvider<List<PublicUser>>(
  (ref) async =>
      unwrapList(await ref.watch(userRepositoryProvider).following()),
  retry: noRetry,
);

/// Suggested channels.
final recommendedChannelsProvider = FutureProvider<List<PublicUser>>(
  (ref) async =>
      unwrapList(await ref.watch(userRepositoryProvider).recommended()),
  retry: noRetry,
);
