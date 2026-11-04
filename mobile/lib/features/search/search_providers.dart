/// Riverpod providers for search. Every provider is keyed by the (trimmed)
/// query and `autoDispose`: a new query reads a fresh provider, so an answer
/// to an older query can never show up, and leaving the signed-in shell drops
/// everything.
///
/// The All tab uses one preview provider per type (so one failing group does
/// not hide the others); a type tab uses a paged notifier of the same type.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/search/data/search_repository.dart';
import 'package:letslive_mobile/features/search/domain/search_type.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// The search endpoints.
final searchRepositoryProvider = Provider<SearchRepository>(
  (ref) => SearchRepository(ref.watch(apiClientProvider)),
);

/// First [searchPreviewLimit] channels for a query (All tab).
final usersPreviewProvider = FutureProvider.autoDispose
    .family<List<PublicUser>, String>(
      (ref, q) async => unwrapList(
        await ref
            .watch(searchRepositoryProvider)
            .users(q: q, limit: searchPreviewLimit),
      ),
      retry: noRetry,
    );

/// First [searchPreviewLimit] live streams for a query (All tab).
final livesPreviewProvider = FutureProvider.autoDispose
    .family<List<Livestream>, String>(
      (ref, q) async => unwrapList(
        await ref
            .watch(searchRepositoryProvider)
            .lives(q: q, limit: searchPreviewLimit),
      ),
      retry: noRetry,
    );

/// First [searchPreviewLimit] videos for a query (All tab).
final vodsPreviewProvider = FutureProvider.autoDispose
    .family<List<Vod>, String>(
      (ref, q) async => unwrapList(
        await ref
            .watch(searchRepositoryProvider)
            .vods(q: q, limit: searchPreviewLimit),
      ),
      retry: noRetry,
    );

/// Channels for [query], page by page (Channels tab).
class UsersSearchNotifier extends PagedFeedNotifier<PublicUser> {
  UsersSearchNotifier(this.query);

  final String query;

  @override
  int get pageSize => searchPageSize;

  @override
  String idOf(PublicUser item) => item.id;

  @override
  Future<ApiResponse<List<PublicUser>>> fetchPage(int page) =>
      ref.read(searchRepositoryProvider).users(q: query, page: page);
}

/// Live streams for [query], page by page (Live tab).
class LivesSearchNotifier extends PagedFeedNotifier<Livestream> {
  LivesSearchNotifier(this.query);

  final String query;

  @override
  int get pageSize => searchPageSize;

  @override
  String idOf(Livestream item) => item.id;

  @override
  Future<ApiResponse<List<Livestream>>> fetchPage(int page) =>
      ref.read(searchRepositoryProvider).lives(q: query, page: page);
}

/// Videos for [query], page by page (Videos tab).
class VodsSearchNotifier extends PagedFeedNotifier<Vod> {
  VodsSearchNotifier(this.query);

  final String query;

  @override
  int get pageSize => searchPageSize;

  @override
  String idOf(Vod item) => item.id;

  @override
  Future<ApiResponse<List<Vod>>> fetchPage(int page) =>
      ref.read(searchRepositoryProvider).vods(q: query, page: page);
}

/// Paged channel results by query.
final usersSearchProvider = AsyncNotifierProvider.autoDispose
    .family<UsersSearchNotifier, PagedFeed<PublicUser>, String>(
      UsersSearchNotifier.new,
      retry: noRetry,
    );

/// Paged live results by query.
final livesSearchProvider = AsyncNotifierProvider.autoDispose
    .family<LivesSearchNotifier, PagedFeed<Livestream>, String>(
      LivesSearchNotifier.new,
      retry: noRetry,
    );

/// Paged video results by query.
final vodsSearchProvider = AsyncNotifierProvider.autoDispose
    .family<VodsSearchNotifier, PagedFeed<Vod>, String>(
      VodsSearchNotifier.new,
      retry: noRetry,
    );
