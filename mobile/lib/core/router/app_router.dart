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
import '../../features/settings/presentation/chat_commands_settings_screen.dart';
import '../../features/settings/presentation/profile_settings_screen.dart';
import '../../features/settings/presentation/security_settings_screen.dart';
import '../../features/settings/presentation/stream_settings_screen.dart';
import '../../features/settings/presentation/vods_settings_screen.dart';
import '../../features/livestream/presentation/livestream_screen.dart';
import '../../features/messages/presentation/conversation_screen.dart';
import '../../features/vod/presentation/upload_vod_screen.dart';
import '../../features/vod/presentation/vod_player_screen.dart';
import '../../features/wallet/presentation/wallet_screen.dart';
import '../../features/wallet/presentation/wallet_transactions_screen.dart';
import '../../features/wallet/presentation/wallet_deposit_screen.dart';
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
  static const settingsChatCommands = '/settings/chat-commands';
  static const settingsVods = '/settings/vods';
  static const search = '/search';
  static String userProfile(String userId) => '/users/$userId';
  static String livestream(String userId) => '/livestream/$userId';
  static String vodPlayer(String vodId) => '/vods/$vodId/watch';
  static String conversation(String id) => '/conversations/$id';
  static const uploadVod = '/upload-vod';
  static const wallet = '/wallet';
  static const walletTransactions = '/wallet/transactions';
  static const walletDeposit = '/wallet/deposit';
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
          pageBuilder: (context, state) =>
              const NoTransitionPage(child: HomeScreen()),
        ),
        GoRoute(
          path: AppRoutes.messages,
          pageBuilder: (context, state) =>
              const NoTransitionPage(child: MessagesScreen()),
        ),
        GoRoute(
          path: AppRoutes.notifications,
          pageBuilder: (context, state) =>
              const NoTransitionPage(child: NotificationsScreen()),
        ),
        GoRoute(
          path: AppRoutes.settings,
          pageBuilder: (context, state) =>
              const NoTransitionPage(child: SettingsScreen()),
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
      path: AppRoutes.settingsChatCommands,
      redirect: _requireAuth,
      builder: (context, state) => const ChatCommandsSettingsScreen(),
    ),
    GoRoute(
      path: AppRoutes.settingsVods,
      redirect: _requireAuth,
      builder: (context, state) => const VodsSettingsScreen(),
    ),
    GoRoute(
      path: AppRoutes.uploadVod,
      redirect: _requireAuth,
      builder: (context, state) => const UploadVodScreen(),
    ),

    // Wallet screens (outside shell for full-screen view)
    GoRoute(
      path: AppRoutes.wallet,
      redirect: _requireAuth,
      builder: (context, state) => const WalletScreen(),
    ),
    GoRoute(
      path: AppRoutes.walletTransactions,
      redirect: _requireAuth,
      builder: (context, state) => const WalletTransactionsScreen(),
    ),
    GoRoute(
      path: AppRoutes.walletDeposit,
      redirect: _requireAuth,
      builder: (context, state) => const WalletDepositScreen(),
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

    // Livestream viewer (video player + live chat)
    GoRoute(
      path: '/livestream/:userId',
      builder: (context, state) {
        final userId = state.pathParameters['userId']!;
        final livestreamId = state.uri.queryParameters['livestreamId'] ?? '';
        return LivestreamScreen(userId: userId, livestreamId: livestreamId);
      },
    ),

    // VOD player
    GoRoute(
      path: '/vods/:vodId/watch',
      builder: (context, state) {
        final vodId = state.pathParameters['vodId']!;
        return VodPlayerScreen(vodId: vodId);
      },
    ),

    // Conversation detail (DM chat)
    GoRoute(
      path: '/conversations/:conversationId',
      redirect: _requireAuth,
      builder: (context, state) {
        final conversationId = state.pathParameters['conversationId']!;
        return ConversationScreen(conversationId: conversationId);
      },
    ),
  ],
);
