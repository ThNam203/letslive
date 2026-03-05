import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';

import '../config/app_config.dart';
import 'api_endpoints.dart';
import 'api_response.dart';

class ApiClient {
  late final Dio _dio;
  late final CookieJar _cookieJar;
  Completer<void>? _refreshCompleter;

  ApiClient() {
    _cookieJar = CookieJar();

    _dio = Dio(
      BaseOptions(
        baseUrl: AppConfig.apiUrl,
        connectTimeout: AppConfig.requestTimeout,
        receiveTimeout: AppConfig.requestTimeout,
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 'no-store',
          'User-Agent': 'letslive-mobile/1.0 (Dart/Flutter)',
        },
      ),
    );

    _dio.interceptors.addAll([
      CookieManager(_cookieJar),
      _AuthInterceptor(this),
    ]);
  }

  Dio get dio => _dio;
  CookieJar get cookieJar => _cookieJar;

  /// Perform a token refresh, deduplicating concurrent requests.
  Future<void> refreshToken() async {
    if (_refreshCompleter != null) {
      return _refreshCompleter!.future;
    }

    _refreshCompleter = Completer<void>();

    try {
      await _dio.post(ApiEndpoints.authRefreshToken);
      _refreshCompleter!.complete();
    } catch (e) {
      _refreshCompleter!.completeError(e);
    } finally {
      _refreshCompleter = null;
    }
  }

  /// Clear all cookies (used on logout).
  Future<void> clearCookies() async {
    _cookieJar.deleteAll();
  }

  // ── Convenience methods ────────────────────────────────────────

  Future<ApiResponse<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    T Function(dynamic)? fromJsonT,
  }) async {
    final response = await _dio.get(path, queryParameters: queryParameters);
    return ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      response.statusCode ?? 0,
      fromJsonT: fromJsonT,
    );
  }

  Future<ApiResponse<T>> post<T>(
    String path, {
    dynamic data,
    T Function(dynamic)? fromJsonT,
  }) async {
    final response = await _dio.post(path, data: data);
    return ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      response.statusCode ?? 0,
      fromJsonT: fromJsonT,
    );
  }

  Future<ApiResponse<T>> put<T>(
    String path, {
    dynamic data,
    T Function(dynamic)? fromJsonT,
  }) async {
    final response = await _dio.put(path, data: data);
    return ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      response.statusCode ?? 0,
      fromJsonT: fromJsonT,
    );
  }

  Future<ApiResponse<T>> patch<T>(
    String path, {
    dynamic data,
    T Function(dynamic)? fromJsonT,
  }) async {
    final response = await _dio.patch(path, data: data);
    return ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      response.statusCode ?? 0,
      fromJsonT: fromJsonT,
    );
  }

  Future<ApiResponse<T>> delete<T>(
    String path, {
    T Function(dynamic)? fromJsonT,
  }) async {
    final response = await _dio.delete(path);
    return ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      response.statusCode ?? 0,
      fromJsonT: fromJsonT,
    );
  }

  Future<ApiResponse<T>> upload<T>(
    String path, {
    required FormData formData,
    T Function(dynamic)? fromJsonT,
  }) async {
    final response = await _dio.post(
      path,
      data: formData,
      options: Options(headers: {'Content-Type': 'multipart/form-data'}),
    );
    return ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      response.statusCode ?? 0,
      fromJsonT: fromJsonT,
    );
  }
}

class _AuthInterceptor extends Interceptor {
  final ApiClient _client;

  _AuthInterceptor(this._client);

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401 &&
        !_shouldSkipRefresh(err.requestOptions.path)) {
      try {
        await _client.refreshToken();

        // Retry the original request
        final response = await _client.dio.fetch(err.requestOptions);
        return handler.resolve(response);
      } catch (_) {
        return handler.next(err);
      }
    }

    handler.next(err);
  }

  bool _shouldSkipRefresh(String path) {
    return ApiEndpoints.refreshExcludePaths.any(
      (excluded) => path.startsWith(excluded),
    );
  }
}
