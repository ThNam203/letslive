import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';

/// Videos per page of `GET /vods/author` (the backend default and maximum).
const int myVodsPageSize = 20;

/// The creator's own videos: list, edit, delete and upload (web
/// `GetAllVODsAsAuthor`, `UpdateVOD`, `DeleteVOD`, `UploadVOD` in
/// `web/lib/api/vod.ts`). Failures are values, never thrown.
class MyVodsRepository {
  const MyVodsRepository(this._client);

  final ApiClient _client;

  /// `GET /vods/author`: the user's videos of any status, [page] 0-based.
  Future<ApiResponse<List<Vod>>> list(int page) => _client.get(
    ApiEndpoints.myVods,
    query: {'page': page, 'limit': myVodsPageSize},
    parse: Vod.parseList,
  );

  /// `PATCH /vods/<id>`; [thumbnailPath] (from `/upload-file`) is sent only
  /// when a new thumbnail was chosen.
  Future<ApiResponse<Object?>> update(
    String id, {
    required String title,
    required String description,
    required bool isPublic,
    String? thumbnailPath,
  }) => _client.patch(
    '${ApiEndpoints.vods}/$id',
    body: {
      'title': title,
      'description': description,
      'visibility': isPublic ? 'public' : 'private',
      'thumbnailUrl': ?thumbnailPath,
    },
    parse: (data) => data,
  );

  /// `DELETE /vods/<id>`.
  Future<ApiResponse<Object?>> delete(String id) =>
      _client.delete('${ApiEndpoints.vods}/$id', parse: (data) => data);

  /// `POST /vods/upload`: sends the file at [path] (streamed from disk, not
  /// loaded into memory) with its details. Answers the new video, status
  /// `processing`. A fresh form per call, so a retry re-sends everything.
  Future<ApiResponse<Vod>> upload({
    required String path,
    required String fileName,
    required String title,
    required String description,
    required bool isPublic,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancel,
  }) async => _client.upload(
    ApiEndpoints.uploadVod,
    formData: FormData.fromMap({
      'file': await MultipartFile.fromFile(path, filename: fileName),
      'title': title,
      'description': description,
      'visibility': isPublic ? 'public' : 'private',
    }),
    onProgress: onProgress,
    cancel: cancel,
    parse: Vod.parse,
  );
}
