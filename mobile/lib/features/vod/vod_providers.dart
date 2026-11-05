/// Riverpod providers for the VOD screen (`/users/:userId/vods/:vodId`):
/// the video, the viewer's reaction to it and the uploader's other videos.
/// Comments live in `comment_providers.dart`.
///
/// All data providers are `autoDispose`: leaving the screen (or logging out)
/// drops them, so the next account never sees the previous one's data.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/vod/data/vod_reaction.dart';
import 'package:letslive_mobile/features/vod/data/vod_repository.dart';

/// Video endpoints.
final vodRepositoryProvider = Provider<VodRepository>(
  (ref) => VodRepository(ref.watch(apiClientProvider)),
);

/// One video, or null when it does not exist (404), is private to someone
/// else (also 404) or the id is malformed (400): the screen's not-found
/// state. Other failures are the provider's error, with a retry.
final vodProvider = FutureProvider.autoDispose.family<Vod?, String>((
  ref,
  vodId,
) async {
  final response = await ref.watch(vodRepositoryProvider).byId(vodId);
  if (response.statusCode == 404 || response.statusCode == 400) return null;
  return unwrapData(response);
}, retry: noRetry);

/// The public videos of one user (first page), for "Other streams".
final userVodsProvider = FutureProvider.autoDispose.family<List<Vod>, String>(
  (ref, userId) async =>
      unwrapList(await ref.watch(vodRepositoryProvider).ofUser(userId)),
  retry: noRetry,
);

/// The signed-in viewer's reaction to a video, with the like count.
final vodReactionProvider = AsyncNotifierProvider.autoDispose
    .family<VodReactionNotifier, VodReactionState, String>(
      VodReactionNotifier.new,
      retry: noRetry,
    );

/// Loads the viewer's reaction and changes it. Like web, a change shows at
/// once and is undone if the server refuses.
class VodReactionNotifier extends AsyncNotifier<VodReactionState> {
  VodReactionNotifier(this.vodId);

  /// The video reacted to.
  final String vodId;

  bool _isPending = false;

  @override
  Future<VodReactionState> build() async =>
      unwrapData(await ref.watch(vodRepositoryProvider).myReaction(vodId));

  /// Taps [clicked]: sets it, or removes it when it already is the viewer's
  /// reaction. [current] is what the buttons show now (the loaded state, or
  /// the video's own like count while the reaction is still loading).
  /// Ignored while a change is in flight, like web. Returns the failed
  /// response when the server refused (the buttons are back to [current]),
  /// else null.
  Future<ApiResponse<Object?>?> react(
    VodReaction clicked,
    VodReactionState current,
  ) async {
    if (_isPending) return null;
    _isPending = true;
    final next = current.reaction == clicked ? null : clicked;
    state = AsyncData(current.reactedWith(next));
    final repo = ref.read(vodRepositoryProvider);
    final response = await (next == null
        ? repo.removeReaction(vodId)
        : repo.react(vodId, next));
    _isPending = false;
    if (!ref.mounted) return null;
    final server = response.data;
    if (response.success && server != null) {
      state = AsyncData(server);
      return null;
    }
    state = AsyncData(current);
    return response;
  }
}
