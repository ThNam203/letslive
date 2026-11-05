import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/creator/data/my_vods_repository.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/vod/data/vod_repository.dart';

import '../home/home_fixtures.dart';

/// Fakes of the creator repositories shared by the manager and screen tests.

Vod vodWithStatus(String status) =>
    Vod.fromJson({...vodJson(id: 'vod-9'), 'status': status});

ApiResponse<T> okResponse<T>(T data) => ApiResponse<T>(
  success: true,
  statusCode: 200,
  code: 0,
  key: 'ok',
  message: 'ok',
  requestId: 'r',
  data: data,
);

ApiResponse<T> failResponse<T>(int status) => ApiResponse<T>(
  success: false,
  statusCode: status,
  code: 1,
  key: 'res_err_internal_server',
  message: 'no',
  requestId: 'r',
);

/// Records the uploads it is asked for; [onUpload] decides what they do.
class FakeMyVods implements MyVodsRepository {
  FakeMyVods(this.onUpload);

  Future<ApiResponse<Vod>> Function(
    void Function(int, int)? onProgress,
    CancelToken? cancel,
  )
  onUpload;
  final List<String> uploadedPaths = [];

  @override
  Future<ApiResponse<Vod>> upload({
    required String path,
    required String fileName,
    required String title,
    required String description,
    required bool isPublic,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancel,
  }) {
    uploadedPaths.add(path);
    return onUpload(onProgress, cancel);
  }

  @override
  Future<ApiResponse<List<Vod>>> list(int page) async => okResponse(<Vod>[]);

  @override
  Future<ApiResponse<Object?>> update(
    String id, {
    required String title,
    required String description,
    required bool isPublic,
    String? thumbnailPath,
  }) async => okResponse(null);

  @override
  Future<ApiResponse<Object?>> delete(String id) async => okResponse(null);
}

/// Answers the status polls from [statuses], the last one repeating.
class FakeVodStatus implements VodRepository {
  FakeVodStatus(this.statuses);

  List<String> statuses;
  int polls = 0;

  @override
  Future<ApiResponse<Vod>> byId(String vodId) async {
    final status =
        statuses[polls < statuses.length ? polls : statuses.length - 1];
    polls++;
    return okResponse(vodWithStatus(status));
  }

  @override
  dynamic noSuchMethod(Invocation invocation) =>
      throw UnimplementedError('${invocation.memberName}');
}
