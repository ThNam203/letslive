import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/network/api_client.dart';
import 'features/auth/data/auth_repository.dart';
import 'models/user.dart';

/// Global API client singleton.
final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});

/// Auth repository.
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(ref.watch(apiClientProvider));
});

/// Current authenticated user (null when not logged in).
final currentUserProvider =
    NotifierProvider<CurrentUserNotifier, User?>(CurrentUserNotifier.new);

class CurrentUserNotifier extends Notifier<User?> {
  @override
  User? build() => null;

  AuthRepository get _authRepository => ref.read(authRepositoryProvider);

  Future<void> fetchMe() async {
    try {
      final response = await _authRepository.getMe();
      if (response.success && response.data != null) {
        state = response.data;
      }
    } catch (_) {
      state = null;
    }
  }

  void setUser(User? user) {
    state = user;
  }

  Future<void> logout() async {
    await _authRepository.logout();
    state = null;
  }
}
