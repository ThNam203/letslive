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

/// One result type of a query, page by page. [_fetch] says which endpoint
/// and [_idOf] identifies an item; the rest is `PagedFeedNotifier`.
class SearchNotifier<T> extends PagedFeedNotifier<T> {
  SearchNotifier(this.query, this._fetch, this._idOf);

  /// The trimmed text being searched.
  final String query;
  final Future<ApiResponse<List<T>>> Function(
    SearchRepository repository,
    String q,
    int page,
  )
  _fetch;
  final String Function(T item) _idOf;

  @override
  int get pageSize => searchPageSize;

  @override
  String idOf(T item) => _idOf(item);

  @override
  Future<ApiResponse<List<T>>> fetchPage(int page) =>
      _fetch(ref.read(searchRepositoryProvider), query, page);
}

/// Paged channel results by query (Channels tab).
final usersSearchProvider = AsyncNotifierProvider.autoDispose
    .family<SearchNotifier<PublicUser>, PagedFeed<PublicUser>, String>(
      (q) => SearchNotifier(
        q,
        (r, q, page) => r.users(q: q, page: page),
        (u) => u.id,
      ),
      retry: noRetry,
    );

/// Paged live results by query (Live tab).
final livesSearchProvider = AsyncNotifierProvider.autoDispose
    .family<SearchNotifier<Livestream>, PagedFeed<Livestream>, String>(
      (q) => SearchNotifier(
        q,
        (r, q, page) => r.lives(q: q, page: page),
        (s) => s.id,
      ),
      retry: noRetry,
    );

/// Paged video results by query (Videos tab).
final vodsSearchProvider = AsyncNotifierProvider.autoDispose
    .family<SearchNotifier<Vod>, PagedFeed<Vod>, String>(
      (q) => SearchNotifier(
        q,
        (r, q, page) => r.vods(q: q, page: page),
        (v) => v.id,
      ),
      retry: noRetry,
    );
