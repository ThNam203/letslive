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
  }) {
    return _client.post(
      ApiEndpoints.authLogin,
      data: {
        'email': email,
        'password': password,
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
      data: {
        'oldPassword': oldPassword,
        'newPassword': newPassword,
      },
    );
  }

  Future<ApiResponse<void>> requestVerification({
    required String email,
  }) {
    return _client.post(
      ApiEndpoints.authVerifyEmail,
      data: {'email': email},
    );
  }

  Future<ApiResponse<void>> verifyOtp({
    required String email,
    required String otpCode,
  }) {
    return _client.post(
      ApiEndpoints.authVerifyOtp,
      data: {
        'email': email,
        'otpCode': otpCode,
      },
    );
  }

  Future<ApiResponse<User>> getMe() {
    return _client.get(
      ApiEndpoints.userMe,
      fromJsonT: (json) => User.fromJson(json as Map<String, dynamic>),
    );
  }
}
