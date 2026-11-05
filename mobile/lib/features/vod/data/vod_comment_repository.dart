import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/vod/data/vod_comment.dart';

/// Comments per page (web `COMMENTS_PAGE_SIZE`).
const int commentsPageSize = 10;

/// Replies per page (web loads 20 at a time).
const int repliesPageSize = 20;

/// The comment endpoints, mirroring `web/lib/api/vod-comment.ts`. Failures
/// are values ([ApiResponse.success] false), never thrown.
class VodCommentRepository {
  const VodCommentRepository(this._client);

  final ApiClient _client;

  /// `GET /vods/<id>/comments`: top-level comments, [commentsPageSize] per
  /// [page] (0-based). `meta.total` is the number of comments.
  Future<ApiResponse<List<VodComment>>> list(String vodId, {int page = 0}) =>
      _client.get(
        '${ApiEndpoints.vods}/$vodId/comments',
        query: {'page': page, 'limit': commentsPageSize},
        parse: VodComment.parseList,
      );

  /// `GET /vod-comments/<id>/replies`: replies to a comment, [repliesPageSize]
  /// per [page] (0-based).
  Future<ApiResponse<List<VodComment>>> replies(
    String commentId, {
    int page = 0,
  }) => _client.get(
    '${ApiEndpoints.vodComments}/$commentId/replies',
    query: {'page': page, 'limit': repliesPageSize},
    parse: VodComment.parseList,
  );

  /// `POST /vods/<id>/comments`: a new comment, or a reply when [parentId] is
  /// set. The answer has no author; the caller adds it.
  Future<ApiResponse<VodComment>> create(
    String vodId,
    String content, {
    String? parentId,
  }) => _client.post(
    '${ApiEndpoints.vods}/$vodId/comments',
    body: {'content': content, 'parentId': ?parentId},
    parse: VodComment.parse,
  );

  /// `PATCH /vod-comments/<id>`: replaces the text of an own comment.
  Future<ApiResponse<VodComment>> edit(String commentId, String content) =>
      _client.patch(
        '${ApiEndpoints.vodComments}/$commentId',
        body: {'content': content},
        parse: VodComment.parse,
      );

  /// `DELETE /vod-comments/<id>`: deletes an own comment (the server keeps it
  /// as "deleted" when it has replies).
  Future<ApiResponse<Object?>> delete(String commentId) => _client.delete(
    '${ApiEndpoints.vodComments}/$commentId',
    parse: (data) => data,
  );

  /// `POST /vod-comments/<id>/like`.
  Future<ApiResponse<Object?>> like(String commentId) => _client.post(
    '${ApiEndpoints.vodComments}/$commentId/like',
    parse: (data) => data,
  );

  /// `DELETE /vod-comments/<id>/like`.
  Future<ApiResponse<Object?>> unlike(String commentId) => _client.delete(
    '${ApiEndpoints.vodComments}/$commentId/like',
    parse: (data) => data,
  );

  /// `POST /vod-comments/liked-ids`: which of [commentIds] (1 to 100) the
  /// signed-in user liked.
  Future<ApiResponse<List<String>>> likedIds(List<String> commentIds) =>
      _client.post(
        '${ApiEndpoints.vodComments}/liked-ids',
        body: {'commentIds': commentIds},
        parse: (data) => [for (final id in data as List<Object?>) id as String],
      );
}
