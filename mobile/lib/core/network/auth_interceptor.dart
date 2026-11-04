import 'dart:async';

import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

/// Marks a request that was already replayed after a refresh.
const String retriedAfterRefreshKey = 'retriedAfterRefresh';

/// Paths that never trigger a refresh (mirrors `REFRESH_EXCLUDE_PATHS` in
/// `web/utils/fetchClient.ts`).
const List<String> _refreshExcludedPaths = [
  '/auth/login',
  '/auth/register',
  ApiEndpoints.refreshToken,
  '/auth/logout',
];

bool shouldSkipRefresh(String path) =>
    _refreshExcludedPaths.any(path.startsWith);

/// Single-flight `POST /auth/refresh-token`: concurrent callers share one
/// in-flight refresh. Emits on [sessionExpired] when the server rejects the
/// refresh (4xx); a network error or 5xx is not a session expiry.
class SessionRefresher {
  SessionRefresher(this._dio);

  final Dio _dio;
  final StreamController<void> _expired = StreamController<void>.broadcast();
  Future<bool>? _inFlight;

  Stream<void> get sessionExpired => _expired.stream;

  Future<bool> refresh() => _inFlight ??= _doRefresh().whenComplete(() {
    _inFlight = null;
  });

  Future<bool> _doRefresh() async {
    try {
      final response = await _dio.post<String>(ApiEndpoints.refreshToken);
      final status = response.statusCode ?? 0;
      if (status >= 200 && status < 300) return true;
      if (status >= 400 && status < 500) _expired.add(null);
      return false;
    } on DioException {
      return false;
    }
  }

  Future<void> dispose() => _expired.close();
}

/// On a 401 for a non-auth path: refresh once (shared), then replay the
/// request once. If the refresh fails the original 401 is passed through, so
/// there is no loop.
class AuthInterceptor extends Interceptor {
  AuthInterceptor(this._dio, this._refresher);

  final Dio _dio;
  final SessionRefresher _refresher;

  @override
  Future<void> onResponse(
    Response<dynamic> response,
    ResponseInterceptorHandler handler,
  ) async {
    final options = response.requestOptions;
    if (response.statusCode != 401 ||
        shouldSkipRefresh(options.path) ||
        options.extra[retriedAfterRefreshKey] == true) {
      handler.next(response);
      return;
    }

    if (!await _refresher.refresh()) {
      handler.next(response);
      return;
    }

    try {
      final body = options.data;
      final retry = options.copyWith(
        // A FormData stream can only be consumed once.
        data: body is FormData ? body.clone() : body,
        extra: {...options.extra, retriedAfterRefreshKey: true},
      );
      handler.resolve(await _dio.fetch<dynamic>(retry));
    } on DioException catch (e) {
      handler.reject(e);
    }
  }
}
