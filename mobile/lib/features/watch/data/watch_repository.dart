import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/home/data/livestream.dart';

/// Live stream lookups for the channel screen, mirroring
/// `GetLivestreamOfUser` in `web/lib/api/livestream.ts`. Failures are values
/// ([ApiResponse.success] false), never thrown.
class WatchRepository {
  const WatchRepository(this._client);

  final ApiClient _client;

  /// `GET /livestreams?userId=<id>`: the stream [userId] is broadcasting.
  /// A user who is not live is a success with null data (the backend sends
  /// `data: null`), not an error. The backend does not check that the user
  /// exists; the profile lookup does.
  Future<ApiResponse<Livestream>> liveOf(String userId) => _client.get(
    ApiEndpoints.livestreams,
    query: {'userId': userId},
    parse: Livestream.parse,
  );
}
