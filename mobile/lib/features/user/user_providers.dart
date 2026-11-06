/// Riverpod providers for public user data. They are `autoDispose`, so
/// leaving the signed-in shell (logout) drops them and the next account never
/// sees the previous one's data.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/data/upload_file_repository.dart';
import 'package:letslive_mobile/features/user/data/user_repository.dart';

/// Public user lookups.
final userRepositoryProvider = Provider<UserRepository>(
  (ref) => UserRepository(ref.watch(apiClientProvider)),
);

/// The general file upload (`/upload-file`): DM pictures, VOD thumbnails.
final uploadFileRepositoryProvider = Provider<UploadFileRepository>(
  (ref) => UploadFileRepository(ref.watch(apiClientProvider)),
);

/// One user by id. Cards of the same owner share one request (like web's
/// per-id query cache); invalidate the family to reload.
final publicUserProvider = AsyncNotifierProvider.autoDispose
    .family<PublicUserNotifier, PublicUser, String>(
      PublicUserNotifier.new,
      retry: noRetry,
    );

/// Loads one user and follows or unfollows them.
class PublicUserNotifier extends AsyncNotifier<PublicUser> {
  PublicUserNotifier(this.userId);

  /// The user shown.
  final String userId;

  bool _isPending = false;

  @override
  Future<PublicUser> build() async =>
      unwrapData(await ref.watch(userRepositoryProvider).byId(userId));

  /// Follows the user, or unfollows when the viewer already does. The
  /// follower count and the button change at once; if the server refuses,
  /// both go back. Ignored while a request is in flight (double taps).
  /// Returns the failed response, or null on success or when ignored.
  Future<ApiResponse<Object?>?> toggleFollow() async {
    final current = state.value;
    if (current == null || _isPending) return null;
    _isPending = true;
    final follow = !(current.isFollowing ?? false);
    state = AsyncData(current.withFollowing(following: follow));
    final repo = ref.read(userRepositoryProvider);
    final response = await (follow
        ? repo.follow(userId)
        : repo.unfollow(userId));
    _isPending = false;
    if (!ref.mounted) return null;
    if (!response.success) {
      state = AsyncData(current);
      return response;
    }
    // The profile's People tab may keep this list alive, so tell it the
    // followed channels changed.
    ref.invalidate(followingChannelsProvider);
    return null;
  }
}

/// Channels the signed-in user follows.
final followingChannelsProvider = FutureProvider.autoDispose<List<PublicUser>>(
  (ref) async =>
      unwrapList(await ref.watch(userRepositoryProvider).following()),
  retry: noRetry,
);

/// Suggested channels.
final recommendedChannelsProvider =
    FutureProvider.autoDispose<List<PublicUser>>(
      (ref) async =>
          unwrapList(await ref.watch(userRepositoryProvider).recommended()),
      retry: noRetry,
    );
