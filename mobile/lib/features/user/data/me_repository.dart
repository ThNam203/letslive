import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// The signed-in user's own profile (`/user/me`), mirroring `GetMeProfile`,
/// `UpdateProfile` and the picture uploads in `web/lib/api/user.ts`.
class MeRepository {
  const MeRepository(this._client);

  final ApiClient _client;

  /// `GET /user/me`. A 401 means there is no (valid) session.
  Future<ApiResponse<MeUser>> fetch() =>
      _client.get(ApiEndpoints.userMe, parse: MeUser.parse);

  /// `PUT /user/me` with only the given fields (the server leaves omitted
  /// ones alone). Returns the updated user.
  Future<ApiResponse<MeUser>> update({
    String? username,
    String? bio,
    Map<String, String>? socialMediaLinks,
  }) => _client.put(
    ApiEndpoints.userMe,
    body: {
      'username': ?username,
      'bio': ?bio,
      'socialMediaLinks': ?socialMediaLinks,
    },
    parse: MeUser.parse,
  );

  /// `PATCH /user/me/profile-picture`, multipart field `profile-picture`.
  /// Answers the new picture URL.
  Future<ApiResponse<String>> uploadProfilePicture(
    PickedImage image, {
    void Function(int sent, int total)? onProgress,
  }) => _upload(
    ApiEndpoints.profilePicture,
    'profile-picture',
    image,
    onProgress,
  );

  /// `PATCH /user/me/background-picture`, multipart field
  /// `background-picture`. Answers the new picture URL.
  Future<ApiResponse<String>> uploadBackgroundPicture(
    PickedImage image, {
    void Function(int sent, int total)? onProgress,
  }) => _upload(
    ApiEndpoints.backgroundPicture,
    'background-picture',
    image,
    onProgress,
  );

  Future<ApiResponse<String>> _upload(
    String path,
    String field,
    PickedImage image,
    void Function(int sent, int total)? onProgress,
  ) {
    final form = FormData.fromMap({field: image.toMultipart()});
    return _client.upload(
      path,
      formData: form,
      method: 'PATCH',
      onProgress: onProgress,
      parse: (data) => data is String
          ? data
          : throw FormatException('picture url: expected string, got $data'),
    );
  }
}
