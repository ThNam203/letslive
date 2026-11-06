import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';

/// The `/user/me/devices` endpoints (backend `device` handlers): tells the
/// server which FCM token belongs to the signed-in user. Failures are values
/// ([ApiResponse.success] false), never thrown.
class DeviceRepository {
  const DeviceRepository(this._client);

  final ApiClient _client;

  /// `POST /user/me/devices`: stores [token] for the signed-in user. Safe to
  /// repeat (the server upserts and moves the token to this user).
  /// [platform] is `ios` or `android`; [locale] a BCP 47 tag such as `vi`.
  Future<ApiResponse<Object?>> register({
    required String token,
    required String platform,
    required String locale,
  }) => _client.post(
    ApiEndpoints.userDevices,
    body: {'token': token, 'platform': platform, 'locale': locale},
    parse: (data) => data,
  );

  /// `DELETE /user/me/devices/<token>`. The token goes into the path
  /// URL-encoded: FCM tokens contain a colon.
  Future<ApiResponse<Object?>> unregister(String token) => _client.delete(
    '${ApiEndpoints.userDevices}/${Uri.encodeComponent(token)}',
    parse: (data) => data,
  );
}
