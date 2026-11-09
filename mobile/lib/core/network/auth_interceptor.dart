import 'dart:async';

import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

/// Marks a request that was already replayed after a refresh.
const String retriedAfterRefreshKey = 'retriedAfterRefresh';

/// Paths that never trigger a refresh (mirrors `REFRESH_EXCLUDE_PATHS` in
/// `web/utils/fetchClient.ts`).
const List<String> _refreshExcludedPaths = [
  ApiEndpoints.login,
  ApiEndpoints.signup,
  ApiEndpoints.refreshToken,
  ApiEndpoints.logout,
];

bool shouldSkipRefresh(String path) =>
    _refreshExcludedPaths.any(path.startsWith);

/// Statuses meaning the server rejected the refresh token itself. Everything
/// else (408, 429, 5xx, network) is transient and must not log the user out.
const Set<int> _authRejectionStatuses = {400, 401, 403};

/// Single-flight `POST /auth/refresh-token`: concurrent callers share one
/// in-flight refresh. Emits on [sessionExpired] only when the server rejects
/// the refresh token (400/401/403). Mirrors web `hasRefreshToken()`: with no
/// REFRESH_TOKEN cookie (guest) it makes no request, returns false and emits
/// nothing.
class SessionRefresher {
  SessionRefresher(this._dio, this._hasRefreshToken);

  final Dio _dio;
  final Future<bool> Function() _hasRefreshToken;
  final StreamController<void> _expired = StreamController<void>.broadcast();
  Future<bool>? _inFlight;

  Stream<void> get sessionExpired => _expired.stream;

  Future<bool> refresh() => _inFlight ??= _doRefresh().whenComplete(() {
    _inFlight = null;
  });

  Future<bool> _doRefresh() async {
    if (!await _hasRefreshToken()) return false;
    try {
      final response = await _dio.post<String>(ApiEndpoints.refreshToken);
      final status = response.statusCode ?? 0;
      if (status >= 200 && status < 300) return true;
      if (_authRejectionStatuses.contains(status)) _expired.add(null);
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
