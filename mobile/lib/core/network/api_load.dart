import 'package:letslive_mobile/core/network/api_response.dart';

/// A request behind a screen section failed. Carries the response so the UI
/// can show the localized server error and decide whether to offer a retry.
class ApiLoadException implements Exception {
  const ApiLoadException(this.response);

  /// The failed response (server error envelope or transport failure).
  final ApiResponse<Object?> response;

  @override
  String toString() =>
      'ApiLoadException(${response.statusCode} ${response.key})';
}

/// `retry` for providers behind a screen section: errors show at once with a
/// manual retry button instead of Riverpod's silent backoff retries.
Duration? noRetry(int retryCount, Object error) => null;

/// The list of a successful [response]. Go sends `null` for an empty list, so
/// that is an empty list here. A failed response throws [ApiLoadException],
/// which `AsyncValue` turns into an error state with retry.
List<T> unwrapList<T>(ApiResponse<List<T>> response) {
  if (!response.success) throw ApiLoadException(response);
  return response.data ?? const [];
}

/// The data of a successful [response]; a failed one, or a success without
/// data, throws [ApiLoadException].
T unwrapData<T>(ApiResponse<T> response) {
  final data = response.data;
  if (!response.success || data == null) throw ApiLoadException(response);
  return data;
}
