import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/login_screen.dart';
import '../../features/auth/presentation/signup_screen.dart';
import '../../features/home/presentation/home_screen.dart';
import '../../features/home/presentation/main_shell.dart';
import '../../features/messages/presentation/messages_screen.dart';
import '../../features/notifications/presentation/notifications_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/search/presentation/search_screen.dart';
import '../../features/settings/presentation/settings_screen.dart';
import '../../features/settings/presentation/profile_settings_screen.dart';
import '../../features/settings/presentation/security_settings_screen.dart';
import '../../features/settings/presentation/stream_settings_screen.dart';
import '../../features/settings/presentation/vods_settings_screen.dart';
import '../../providers.dart';

abstract final class AppRoutes {
  static const login = '/login';
  static const signup = '/signup';
  static const home = '/';
  static const messages = '/messages';
  static const notifications = '/notifications';
  static const settings = '/settings';
  static const settingsProfile = '/settings/profile';
  static const settingsSecurity = '/settings/security';
  static const settingsStream = '/settings/stream';
  static const settingsVods = '/settings/vods';
  static const search = '/search';
  static String userProfile(String userId) => '/users/$userId';
}

final rootNavigatorKey = GlobalKey<NavigatorState>();
final shellNavigatorKey = GlobalKey<NavigatorState>();

/// Redirect to login if the user is not authenticated.
String? _requireAuth(BuildContext context, GoRouterState state) {
  final container = ProviderScope.containerOf(context);
  final user = container.read(currentUserProvider);
  if (user == null) return AppRoutes.login;
  return null;
}

final appRouter = GoRouter(
  navigatorKey: rootNavigatorKey,
  initialLocation: AppRoutes.home,
  routes: [
    // Auth routes (no bottom nav)
    GoRoute(
      path: AppRoutes.login,
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: AppRoutes.signup,
      builder: (context, state) => const SignupScreen(),
    ),

    // Main app shell with bottom navigation
    ShellRoute(
      navigatorKey: shellNavigatorKey,
      builder: (context, state, child) => MainShell(child: child),
      routes: [
        GoRoute(
          path: AppRoutes.home,
          pageBuilder: (context, state) => const NoTransitionPage(
            child: HomeScreen(),
          ),
        ),
        GoRoute(
          path: AppRoutes.messages,
          pageBuilder: (context, state) => const NoTransitionPage(
            child: MessagesScreen(),
          ),
        ),
        GoRoute(
          path: AppRoutes.notifications,
          pageBuilder: (context, state) => const NoTransitionPage(
            child: NotificationsScreen(),
          ),
        ),
        GoRoute(
          path: AppRoutes.settings,
          pageBuilder: (context, state) => const NoTransitionPage(
            child: SettingsScreen(),
          ),
        ),
      ],
    ),

    // Settings sub-screens (outside shell for full-screen view)
    // Require authentication — redirect to login if not logged in.
    GoRoute(
      path: AppRoutes.settingsProfile,
      redirect: _requireAuth,
      builder: (context, state) => const ProfileSettingsScreen(),
    ),
    GoRoute(
      path: AppRoutes.settingsSecurity,
      redirect: _requireAuth,
      builder: (context, state) => const SecuritySettingsScreen(),
    ),
    GoRoute(
      path: AppRoutes.settingsStream,
      redirect: _requireAuth,
      builder: (context, state) => const StreamSettingsScreen(),
    ),
    GoRoute(
      path: AppRoutes.settingsVods,
      redirect: _requireAuth,
      builder: (context, state) => const VodsSettingsScreen(),
    ),

    // Search (outside shell for full-screen view)
    GoRoute(
      path: AppRoutes.search,
      builder: (context, state) => const SearchScreen(),
    ),

    // Profile (outside shell for full-screen view)
    GoRoute(
      path: '/users/:userId',
      builder: (context, state) {
        final userId = state.pathParameters['userId']!;
        return ProfileScreen(userId: userId);
      },
    ),
  ],
);
