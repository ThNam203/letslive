import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/livestream.dart';

class LivestreamRepository {
  final ApiClient _client;

  LivestreamRepository(this._client);

  Future<ApiResponse<List<Livestream>>> getLivestreams({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.livestreams,
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Livestream.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<Livestream>>> getPopularLivestreams({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.popularLivestreams,
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Livestream.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
