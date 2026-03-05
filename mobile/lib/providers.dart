import 'dart:ui';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'core/network/api_client.dart';
import 'features/auth/data/auth_repository.dart';
import 'features/livestream/data/livestream_repository.dart';
import 'features/messages/data/message_repository.dart';
import 'features/notifications/data/notification_repository.dart';
import 'features/user/data/user_repository.dart';
import 'features/vod/data/vod_repository.dart';
import 'models/user.dart';

// ---------------------------------------------------------------------------
// SharedPreferences – must be overridden in main().
// ---------------------------------------------------------------------------

final sharedPreferencesProvider = Provider<SharedPreferences>((ref) {
  throw UnimplementedError('Must be overridden in main');
});

// ---------------------------------------------------------------------------
// Theme mode
// ---------------------------------------------------------------------------

enum AppThemeMode { light, dark, system }

final themeModeProvider =
    NotifierProvider<ThemeModeNotifier, AppThemeMode>(ThemeModeNotifier.new);

class ThemeModeNotifier extends Notifier<AppThemeMode> {
  static const _key = 'app_theme_mode';

  @override
  AppThemeMode build() {
    final prefs = ref.watch(sharedPreferencesProvider);
    final value = prefs.getString(_key);
    if (value != null) {
      return AppThemeMode.values.firstWhere(
        (e) => e.name == value,
        orElse: () => AppThemeMode.system,
      );
    }
    return AppThemeMode.system;
  }

  Future<void> setThemeMode(AppThemeMode mode) async {
    final prefs = ref.read(sharedPreferencesProvider);
    await prefs.setString(_key, mode.name);
    state = mode;
  }
}

// ---------------------------------------------------------------------------
// Locale (null = follow system)
// ---------------------------------------------------------------------------

/// Display names for supported languages (always in native form).
const supportedLanguageNames = <String, String>{
  'en': 'English',
  'vi': 'Tiếng Việt',
};

final localeProvider =
    NotifierProvider<LocaleNotifier, Locale?>(LocaleNotifier.new);

class LocaleNotifier extends Notifier<Locale?> {
  static const _key = 'app_locale';

  @override
  Locale? build() {
    final prefs = ref.watch(sharedPreferencesProvider);
    final value = prefs.getString(_key);
    if (value != null) {
      return Locale(value);
    }
    return null;
  }

  Future<void> setLocale(Locale? locale) async {
    final prefs = ref.read(sharedPreferencesProvider);
    if (locale != null) {
      await prefs.setString(_key, locale.languageCode);
    } else {
      await prefs.remove(_key);
    }
    state = locale;
  }
}

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

/// Unread notification count.
final unreadNotificationCountProvider =
    NotifierProvider<UnreadNotificationCountNotifier, int>(
        UnreadNotificationCountNotifier.new);

class UnreadNotificationCountNotifier extends Notifier<int> {
  @override
  int build() => 0;

  Future<void> fetch() async {
    try {
      final repo = ref.read(notificationRepositoryProvider);
      final response = await repo.getUnreadCount();
      if (response.success && response.data != null) {
        state = response.data!;
      }
    } catch (_) {
      // Keep current count on failure
    }
  }

  void decrement() {
    if (state > 0) state--;
  }

  void clear() {
    state = 0;
  }
}

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
