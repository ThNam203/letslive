import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

/// The signed-in user's own profile (`/user/me`), mirroring `GetMeProfile`
/// and `UpdateProfile` in `web/lib/api/user.ts`.
class MeRepository {
  const MeRepository(this._client);

  final ApiClient _client;

  /// `GET /user/me`. A 401 means there is no (valid) session.
  Future<ApiResponse<MeUser>> fetch() =>
      _client.get(ApiEndpoints.userMe, parse: MeUser.parse);

  /// `PUT /user/me {username}` (account setup). Returns the updated user.
  Future<ApiResponse<MeUser>> updateUsername(String username) => _client.put(
    ApiEndpoints.userMe,
    body: {'username': username},
    parse: MeUser.parse,
  );
}
