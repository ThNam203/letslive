import 'dart:convert';

/// Pagination info of an envelope (`meta`), mirrors `web/types/fetch-response.ts`.
class Meta {
  const Meta({this.page, this.pageSize, this.total});

  factory Meta.fromJson(Map<String, Object?> json) => Meta(
    page: _asInt(json['page']),
    pageSize: _asInt(json['page_size']),
    total: _asInt(json['total']),
  );

  final int? page;
  final int? pageSize;
  final int? total;
}

/// Why a response did not come from a well-formed server envelope.
enum ApiFailureKind {
  /// The request timed out. Retryable.
  timeout,

  /// No connection / DNS / TLS failure. Retryable.
  network,

  /// The caller cancelled the request. Not retryable.
  cancelled,

  /// The server answered but the body was not a valid envelope (or `parse`
  /// threw). Not retryable by itself.
  malformedResponse,
}

typedef ParseData<T> = T Function(Object? data);

/// Result of every [ApiClient] call. Server error envelopes and transport
/// failures are values (`success == false`), never exceptions.
class ApiResponse<T> {
  const ApiResponse({
    required this.success,
    required this.statusCode,
    required this.code,
    required this.key,
    required this.message,
    required this.requestId,
    this.data,
    this.meta,
    this.errorDetails,
    this.failure,
  });

  /// Transport-level failure: no envelope was received. `statusCode` is 0.
  factory ApiResponse.transportFailure(ApiFailureKind kind, String message) =>
      ApiResponse<T>(
        success: false,
        statusCode: 0,
        code: 0,
        key: 'client.${kind.name}',
        message: message,
        requestId: '',
        failure: kind,
      );

  /// Builds a response from the raw HTTP [body]. [parse] converts the
  /// envelope's `data` into [T]. Never throws.
  factory ApiResponse.fromHttp({
    required int statusCode,
    required String? body,
    required String headerRequestId,
    required ParseData<T> parse,
  }) {
    final ok = statusCode >= 200 && statusCode < 300;
    final trimmed = body?.trim() ?? '';
    if (trimmed.isEmpty) {
      if (ok) {
        return ApiResponse<T>(
          success: true,
          statusCode: statusCode,
          code: 0,
          key: '',
          message: '',
          requestId: headerRequestId,
        );
      }
      return ApiResponse<T>._malformed(
        statusCode,
        headerRequestId,
        'Empty response body',
      );
    }

    final Object? json;
    try {
      json = jsonDecode(trimmed);
    } on FormatException catch (e) {
      return ApiResponse<T>._malformed(
        statusCode,
        headerRequestId,
        'Response is not JSON: ${e.message}',
      );
    }
    if (json is! Map<String, Object?>) {
      return ApiResponse<T>._malformed(
        statusCode,
        headerRequestId,
        'Response is not a JSON object',
      );
    }

    final success = json['success'];
    if (success is! bool) {
      return ApiResponse<T>._malformed(
        statusCode,
        headerRequestId,
        'Envelope has no boolean "success"',
      );
    }

    final requestId = json['requestId'];
    final key = json['key'];
    final message = json['message'];
    final rawData = json['data'];
    final rawMeta = json['meta'];
    T? data;
    if (success && rawData != null) {
      try {
        data = parse(rawData);
      } catch (e) {
        return ApiResponse<T>._malformed(
          statusCode,
          headerRequestId,
          'Could not parse response data: $e',
        );
      }
    }

    return ApiResponse<T>(
      success: success,
      statusCode: statusCode,
      code: _asInt(json['code']) ?? 0,
      key: key is String ? key : '',
      message: message is String ? message : '',
      requestId: requestId is String && requestId.isNotEmpty
          ? requestId
          : headerRequestId,
      data: data,
      meta: rawMeta is Map<String, Object?> ? Meta.fromJson(rawMeta) : null,
      errorDetails: _parseErrorDetails(json['errorDetails']),
    );
  }

  factory ApiResponse._malformed(
    int statusCode,
    String requestId,
    String message,
  ) => ApiResponse<T>(
    success: false,
    statusCode: statusCode,
    code: 0,
    key: 'client.${ApiFailureKind.malformedResponse.name}',
    message: message,
    requestId: requestId,
    failure: ApiFailureKind.malformedResponse,
  );

  final bool success;

  /// HTTP status; 0 when no response was received.
  final int statusCode;

  /// Business-level code.
  final int code;

  /// i18n key.
  final String key;

  /// English default message.
  final String message;
  final T? data;
  final Meta? meta;
  final List<Map<String, Object?>>? errorDetails;
  final String requestId;

  /// Set when this response was synthesized on the client (see
  /// [ApiFailureKind]); null for any real server envelope.
  final ApiFailureKind? failure;

  /// Whether showing a "retry" action can reasonably succeed: timeouts,
  /// connectivity loss and gateway/availability errors (502/503/504).
  bool get isRetryable =>
      failure == ApiFailureKind.timeout ||
      failure == ApiFailureKind.network ||
      statusCode == 502 ||
      statusCode == 503 ||
      statusCode == 504;
}

int? _asInt(Object? value) => value is num ? value.toInt() : null;

List<Map<String, Object?>>? _parseErrorDetails(Object? raw) {
  if (raw is! List<Object?>) return null;
  return [
    for (final item in raw)
      if (item is Map<String, Object?>) item,
  ];
}
