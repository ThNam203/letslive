import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// Public user lookups, mirroring `web/lib/api/user.ts`. Failures are values
/// ([ApiResponse.success] false), never thrown.
class UserRepository {
  const UserRepository(this._client);

  final ApiClient _client;

  /// `GET /user/<id>`: one user's public profile.
  Future<ApiResponse<PublicUser>> byId(String id) =>
      _client.get('${ApiEndpoints.user}/$id', parse: PublicUser.parse);

  /// `GET /user/me/following`: channels the signed-in user follows.
  Future<ApiResponse<List<PublicUser>>> following() =>
      _client.get(ApiEndpoints.following, parse: PublicUser.parseList);

  /// `GET /users/recommendations`: suggested channels (first page).
  Future<ApiResponse<List<PublicUser>>> recommended() => _client.get(
    ApiEndpoints.recommendedChannels,
    query: {'page': 0},
    parse: PublicUser.parseList,
  );
}
