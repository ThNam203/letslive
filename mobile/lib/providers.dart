import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/network/api_client.dart';
import 'features/auth/data/auth_repository.dart';
import 'features/livestream/data/livestream_repository.dart';
import 'features/messages/data/message_repository.dart';
import 'features/notifications/data/notification_repository.dart';
import 'features/user/data/user_repository.dart';
import 'features/vod/data/vod_repository.dart';
import 'models/user.dart';

/// Global API client singleton.
final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});

/// Auth repository.
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(ref.watch(apiClientProvider));
});

/// User repository.
final userRepositoryProvider = Provider<UserRepository>((ref) {
  return UserRepository(ref.watch(apiClientProvider));
});

/// Livestream repository.
final livestreamRepositoryProvider = Provider<LivestreamRepository>((ref) {
  return LivestreamRepository(ref.watch(apiClientProvider));
});

/// VOD repository.
final vodRepositoryProvider = Provider<VodRepository>((ref) {
  return VodRepository(ref.watch(apiClientProvider));
});

/// Notification repository.
final notificationRepositoryProvider = Provider<NotificationRepository>((ref) {
  return NotificationRepository(ref.watch(apiClientProvider));
});

/// Message repository.
final messageRepositoryProvider = Provider<MessageRepository>((ref) {
  return MessageRepository(ref.watch(apiClientProvider));
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
