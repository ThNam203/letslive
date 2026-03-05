import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/vod.dart';

class VodRepository {
  final ApiClient _client;

  VodRepository(this._client);

  Future<ApiResponse<List<Vod>>> getVods({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.vods,
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Vod.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<Vod>> getVod(String id) {
    return _client.get(
      ApiEndpoints.vodById(id),
      fromJsonT: (json) => Vod.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<List<Vod>>> getUserVods(
    String userId, {
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.vods,
      queryParameters: {
        'userId': userId,
        'page': page,
        'limit': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Vod.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<Vod>>> getAuthorVods({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.vodsAuthor,
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Vod.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<Vod>>> getPopularVods({
    int page = 0,
    int pageSize = 20,
  }) {
    return _client.get(
      ApiEndpoints.popularVods,
      queryParameters: {
        'page': page,
        'page_size': pageSize,
      },
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Vod.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<void>> updateVod({
    required String vodId,
    required String title,
    required String description,
    required String visibility,
    String? thumbnailUrl,
  }) {
    return _client.patch(
      ApiEndpoints.vodById(vodId),
      data: {
        'title': title,
        'description': description,
        'visibility': visibility,
        // ignore: use_null_aware_elements
        if (thumbnailUrl != null) 'thumbnailUrl': thumbnailUrl,
      },
    );
  }

  Future<ApiResponse<void>> deleteVod(String vodId) {
    return _client.delete(ApiEndpoints.vodById(vodId));
  }
}
