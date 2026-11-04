import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/network/auth_interceptor.dart';

/// Default per-request timeout (mirrors `DEFAULT_TIMEOUT_MS` in the web client).
const Duration defaultRequestTimeout = Duration(seconds: 15);

/// HTTP client for the LetsLive gateway. Cookie based auth, envelope parsing,
/// silent single-flight token refresh. Calls never throw for server errors or
/// transport failures; inspect [ApiResponse.success] / [ApiResponse.failure].
class ApiClient {
  ApiClient({
    required String baseUrl,
    required CookieJar cookieJar,
    HttpClientAdapter? adapter,
    Duration timeout = defaultRequestTimeout,
  }) : _dio = Dio(
         BaseOptions(
           baseUrl: baseUrl,
           connectTimeout: timeout,
           sendTimeout: timeout,
           receiveTimeout: timeout,
           responseType: ResponseType.plain,
           // Error envelopes are values, not exceptions.
           validateStatus: (_) => true,
           headers: {'Cache-Control': 'no-store'},
         ),
       ) {
    if (adapter != null) _dio.httpClientAdapter = adapter;
    final refreshUri = Uri.parse('$baseUrl${ApiEndpoints.refreshToken}');
    _refresher = SessionRefresher(_dio, () async {
      final cookies = await cookieJar.loadForRequest(refreshUri);
      return cookies.any(
        (c) => c.name == 'REFRESH_TOKEN' && c.value.isNotEmpty,
      );
    });
    _dio.interceptors
      ..add(CookieManager(cookieJar))
      ..add(AuthInterceptor(_dio, _refresher));
  }

  final Dio _dio;
  late final SessionRefresher _refresher;

  /// Fires when the server rejected the refresh token (4xx on
  /// `/auth/refresh-token`), i.e. the user must log in again. Emitted once per
  /// failed refresh, not once per request. Broadcast stream.
  Stream<void> get sessionExpired => _refresher.sessionExpired;

  /// Refreshes the session now. Same single-flight as the automatic 401 path:
  /// concurrent callers share one request. Returns true on success.
  Future<bool> refreshSession() => _refresher.refresh();

  Future<ApiResponse<T>> get<T>(
    String path, {
    Map<String, Object?>? query,
    required ParseData<T> parse,
    CancelToken? cancel,
  }) => _send('GET', path, query: query, parse: parse, cancel: cancel);

  Future<ApiResponse<T>> post<T>(
    String path, {
    Object? body,
    Map<String, Object?>? query,
    required ParseData<T> parse,
    CancelToken? cancel,
  }) => _send(
    'POST',
    path,
    body: body,
    query: query,
    parse: parse,
    cancel: cancel,
  );

  Future<ApiResponse<T>> put<T>(
    String path, {
    Object? body,
    Map<String, Object?>? query,
    required ParseData<T> parse,
    CancelToken? cancel,
  }) => _send(
    'PUT',
    path,
    body: body,
    query: query,
    parse: parse,
    cancel: cancel,
  );

  Future<ApiResponse<T>> patch<T>(
    String path, {
    Object? body,
    Map<String, Object?>? query,
    required ParseData<T> parse,
    CancelToken? cancel,
  }) => _send(
    'PATCH',
    path,
    body: body,
    query: query,
    parse: parse,
    cancel: cancel,
  );

  Future<ApiResponse<T>> delete<T>(
    String path, {
    Object? body,
    Map<String, Object?>? query,
    required ParseData<T> parse,
    CancelToken? cancel,
  }) => _send(
    'DELETE',
    path,
    body: body,
    query: query,
    parse: parse,
    cancel: cancel,
  );

  /// Multipart upload (POST by default). No send/receive timeout (web
  /// `disableTimeout`); stop it with [cancel].
  Future<ApiResponse<T>> upload<T>(
    String path, {
    required FormData formData,
    String method = 'POST',
    Map<String, Object?>? query,
    required ParseData<T> parse,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancel,
  }) => _send(
    method,
    path,
    body: formData,
    query: query,
    parse: parse,
    cancel: cancel,
    onProgress: onProgress,
    noTimeout: true,
  );

  Future<void> dispose() async {
    _dio.close();
    await _refresher.dispose();
  }

  Future<ApiResponse<T>> _send<T>(
    String method,
    String path, {
    Object? body,
    Map<String, Object?>? query,
    required ParseData<T> parse,
    CancelToken? cancel,
    void Function(int sent, int total)? onProgress,
    bool noTimeout = false,
  }) async {
    try {
      final response = await _dio.request<String>(
        path,
        data: body,
        queryParameters: query,
        cancelToken: cancel,
        onSendProgress: onProgress,
        options: Options(
          method: method,
          sendTimeout: noTimeout ? Duration.zero : null,
          receiveTimeout: noTimeout ? Duration.zero : null,
        ),
      );
      return ApiResponse<T>.fromHttp(
        statusCode: response.statusCode ?? 0,
        body: response.data,
        headerRequestId: response.headers.value('x-request-id') ?? '',
        parse: parse,
      );
    } on DioException catch (e) {
      return ApiResponse<T>.transportFailure(_kindOf(e), _messageOf(e));
    }
  }

  static ApiFailureKind _kindOf(DioException e) => switch (e.type) {
    DioExceptionType.connectionTimeout ||
    DioExceptionType.sendTimeout ||
    DioExceptionType.receiveTimeout ||
    DioExceptionType.transformTimeout => ApiFailureKind.timeout,
    DioExceptionType.cancel => ApiFailureKind.cancelled,
    DioExceptionType.badCertificate ||
    DioExceptionType.connectionError ||
    DioExceptionType.badResponse ||
    DioExceptionType.unknown => ApiFailureKind.network,
  };

  static String _messageOf(DioException e) => switch (e.type) {
    DioExceptionType.connectionTimeout ||
    DioExceptionType.sendTimeout ||
    DioExceptionType.receiveTimeout ||
    DioExceptionType.transformTimeout => 'Request timed out',
    DioExceptionType.cancel => 'Request cancelled',
    _ => 'Network error: ${e.message ?? e.type.name}',
  };
}
