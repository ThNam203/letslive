import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';

/// Auth endpoints, mirroring `web/lib/api/auth.ts`. On success the server
/// sets the ACCESS_TOKEN/REFRESH_TOKEN cookies, which [ApiClient]'s cookie jar
/// stores. Failures are values ([ApiResponse.success] false), never thrown.
class AuthRepository {
  AuthRepository({required this._client, required this._botProof});

  final ApiClient _client;
  final BotProof _botProof;

  /// `POST /auth/login`, with bot proof.
  Future<ApiResponse<void>> login(String email, String password) =>
      _postWithBotProof(ApiEndpoints.login, {
        'email': email,
        'password': password,
      });

  /// `POST /auth/verify-email`: emails a 6-digit signup code. With bot proof.
  Future<ApiResponse<void>> requestEmailVerification(String email) =>
      _postWithBotProof(ApiEndpoints.verifyEmail, {'email': email});

  /// `POST /auth/signup` with the emailed code. The backend does not run the
  /// bot check here (the code already proves the email round trip).
  Future<ApiResponse<void>> signup({
    required String email,
    required String username,
    required String password,
    required String otpCode,
  }) => _client.post(
    ApiEndpoints.signup,
    body: {
      'email': email,
      'username': username,
      'password': password,
      'otpCode': otpCode,
    },
    parse: _ignoreData,
  );

  /// `POST /auth/google/mobile` with the id token from native Google sign-in.
  Future<ApiResponse<void>> loginWithGoogle(String idToken) => _client.post(
    ApiEndpoints.googleMobile,
    body: {'idToken': idToken},
    parse: _ignoreData,
  );

  /// `DELETE /auth/logout`: the server expires both cookies.
  Future<ApiResponse<void>> logout() =>
      _client.delete(ApiEndpoints.logout, parse: _ignoreData);

  /// `PATCH /auth/password` (answers 204). No bot proof: the backend only
  /// runs that check on login and verify-email.
  Future<ApiResponse<void>> changePassword({
    required String oldPassword,
    required String newPassword,
  }) => _client.patch(
    ApiEndpoints.changePassword,
    body: {'oldPassword': oldPassword, 'newPassword': newPassword},
    parse: _ignoreData,
  );

  Future<ApiResponse<void>> _postWithBotProof(
    String path,
    Map<String, Object?> body,
  ) async {
    final BotProofFields proof;
    try {
      proof = await _botProof.next();
    } on Exception catch (e) {
      // App Check could not mint a token (usually no connectivity); the user
      // can retry, so report it like a network failure.
      return ApiResponse.transportFailure(
        ApiFailureKind.network,
        'App Check token unavailable: $e',
      );
    }
    return _client.post(
      path,
      body: {...body, ...proof.body},
      headers: proof.headers,
      parse: _ignoreData,
    );
  }

  static void _ignoreData(Object? _) {}
}
