import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/search/domain/search_type.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// The three search endpoints, mirroring `web/lib/api/{user,livestream,vod}.ts`.
/// Failures are values ([ApiResponse.success] false), never thrown.
///
/// One method per result type rather than one dispatching `search(type)`:
/// each returns its own model list, so callers need no casts. The server
/// matches ignoring accents and typos (`docs/SEARCH.md`); the app sends the
/// query as typed. [ApiClient] encodes the query string (Vietnamese, `&`).
/// `page` is 0-based; a page shorter than `limit` is the last one.
class SearchRepository {
  const SearchRepository(this._client);

  final ApiClient _client;

  /// `GET /users/search`: channels whose username matches [q].
  Future<ApiResponse<List<PublicUser>>> users({
    required String q,
    int page = 0,
    int limit = searchPageSize,
  }) => _client.get(
    ApiEndpoints.searchUsers,
    query: {'username': q, 'page': page, 'limit': limit},
    parse: PublicUser.parseList,
  );

  /// `GET /livestreams/search`: streams live now matching [q] (title,
  /// description or streamer name).
  Future<ApiResponse<List<Livestream>>> lives({
    required String q,
    int page = 0,
    int limit = searchPageSize,
  }) => _client.get(
    ApiEndpoints.searchLivestreams,
    query: {'q': q, 'page': page, 'limit': limit},
    parse: Livestream.parseList,
  );

  /// `GET /vods/search`: public, ready videos matching [q].
  Future<ApiResponse<List<Vod>>> vods({
    required String q,
    int page = 0,
    int limit = searchPageSize,
  }) => _client.get(
    ApiEndpoints.searchVods,
    query: {'q': q, 'page': page, 'limit': limit},
    parse: Vod.parseList,
  );
}
