import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// The general file upload (`UploadFile` in `web/lib/api/utils.ts`): one file
/// to `POST /upload-file`, answered with the path it was stored under. Used
/// for direct-message pictures and VOD thumbnails; profile and banner
/// pictures have their own endpoints in `MeRepository`.
class UploadFileRepository {
  const UploadFileRepository(this._client);

  final ApiClient _client;

  /// Uploads [image] (multipart field `file`) and answers its stored path.
  /// Failures are values, like every [ApiClient] call.
  Future<ApiResponse<String>> upload(
    PickedImage image, {
    void Function(int sent, int total)? onProgress,
    CancelToken? cancel,
  }) => _client.upload(
    ApiEndpoints.uploadFile,
    formData: FormData.fromMap({'file': image.toMultipart()}),
    onProgress: onProgress,
    cancel: cancel,
    parse: (data) => data is String && data.isNotEmpty
        ? data
        : throw FormatException('upload-file: expected a path, got $data'),
  );
}
