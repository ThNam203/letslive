import '../../../core/network/api_client.dart';
import '../../../core/network/api_endpoints.dart';
import '../../../core/network/api_response.dart';
import '../../../models/user.dart';

class AuthRepository {
  final ApiClient _client;

  AuthRepository(this._client);

  Future<ApiResponse<void>> login({
    required String email,
    required String password,
    String turnstileToken = '',
  }) {
    return _client.post(
      ApiEndpoints.authLogin,
      data: {
        'email': email,
        'password': password,
        'turnstileToken': turnstileToken,
      },
    );
  }

  Future<ApiResponse<void>> signup({
    required String email,
    required String username,
    required String password,
    required String otpCode,
  }) {
    return _client.post(
      ApiEndpoints.authSignup,
      data: {
        'email': email,
        'username': username,
        'password': password,
        'otpCode': otpCode,
      },
    );
  }

  Future<ApiResponse<void>> logout() {
    return _client.delete(ApiEndpoints.authLogout);
  }

  Future<ApiResponse<void>> changePassword({
    required String oldPassword,
    required String newPassword,
  }) {
    return _client.patch(
      ApiEndpoints.authPassword,
      data: {'oldPassword': oldPassword, 'newPassword': newPassword},
    );
  }

  Future<ApiResponse<void>> loginWithGoogle({required String idToken}) {
    return _client.post(
      ApiEndpoints.authGoogleMobile,
      data: {'idToken': idToken},
    );
  }

  Future<ApiResponse<void>> requestVerification({
    required String email,
    String turnstileToken = '',
  }) {
    return _client.post(
      ApiEndpoints.authVerifyEmail,
      data: {'email': email, 'turnstileToken': turnstileToken},
    );
  }

  Future<ApiResponse<User>> getMe() {
    return _client.get(
      ApiEndpoints.userMe,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }
}
