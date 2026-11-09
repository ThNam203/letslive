import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';

/// Page sizes of the home feed, the same as the web app uses.
const int popularLivePageSize = 10;
const int popularVodsPageSize = 20;

/// The home feed's media lists, mirroring `web/lib/api/livestream.ts` and
/// `vod.ts`. Failures are values ([ApiResponse.success] false), never thrown.
/// The web home has one live list (`/popular-livestreams`: public streams
/// that have not ended); there is no separate "live" endpoint.
class HomeRepository {
  const HomeRepository(this._client);

  final ApiClient _client;

  /// `GET /popular-livestreams`: streams that are live now (first page).
  Future<ApiResponse<List<Livestream>>> popularLive() => _client.get(
    ApiEndpoints.popularLivestreams,
    query: {'page': 0, 'limit': popularLivePageSize},
    parse: Livestream.parseList,
  );

  /// `GET /popular-vods`: popular videos, [popularVodsPageSize] per [page]
  /// (0-based).
  Future<ApiResponse<List<Vod>>> popularVods({int page = 0}) => _client.get(
    ApiEndpoints.popularVods,
    query: {'page': page, 'limit': popularVodsPageSize},
    parse: Vod.parseList,
  );
}
