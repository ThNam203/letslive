import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/vod_comment.dart';

class VodCommentRepository {
  final ApiClient _client;

  VodCommentRepository(this._client);

  Future<ApiResponse<List<VodComment>>> getComments(
    String vodId, {
    int page = 0,
    int limit = 10,
  }) {
    return _client.get(
      ApiEndpoints.vodComments(vodId),
      queryParameters: {'page': page, 'limit': limit},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => VodComment.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<VodComment>>> getReplies(
    String commentId, {
    int page = 0,
    int limit = 20,
  }) {
    return _client.get(
      ApiEndpoints.vodCommentReplies(commentId),
      queryParameters: {'page': page, 'limit': limit},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => VodComment.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<VodComment>> createComment(
    String vodId, {
    required String content,
    String? parentId,
  }) {
    return _client.post(
      ApiEndpoints.vodComments(vodId),
      data: {'content': content, 'parentId': ?parentId},
      fromJsonT: (json) => VodComment.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<void>> deleteComment(String commentId) {
    return _client.delete(ApiEndpoints.vodCommentById(commentId));
  }

  Future<ApiResponse<void>> likeComment(String commentId) {
    return _client.post(ApiEndpoints.vodCommentLike(commentId));
  }

  Future<ApiResponse<void>> unlikeComment(String commentId) {
    return _client.delete(ApiEndpoints.vodCommentLike(commentId));
  }

  Future<ApiResponse<List<String>>> getLikedCommentIds(
    List<String> commentIds,
  ) {
    return _client.post(
      ApiEndpoints.vodCommentLikedIds,
      data: {'commentIds': commentIds},
      fromJsonT: (json) =>
          (json as List<dynamic>).map((e) => e as String).toList(),
    );
  }
}
