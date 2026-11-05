import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/router/app_shell.dart';
import 'package:letslive_mobile/core/router/app_tabs.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/session.dart';
import 'package:letslive_mobile/features/auth/presentation/account_setup_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/signup_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/splash_screen.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_screen.dart';
import 'package:letslive_mobile/features/messages/presentation/new_conversation_screen.dart';
import 'package:letslive_mobile/features/settings/presentation/profile_settings_screen.dart';
import 'package:letslive_mobile/features/settings/presentation/security_screen.dart';
import 'package:letslive_mobile/features/settings/presentation/settings_screen.dart';
import 'package:letslive_mobile/features/shop/presentation/gifts_screen.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_screen.dart';
import 'package:letslive_mobile/features/vod/presentation/vod_screen.dart';
import 'package:letslive_mobile/features/wallet/presentation/wallet_screen.dart';
import 'package:letslive_mobile/features/watch/presentation/channel_screen.dart';

export 'package:letslive_mobile/core/router/app_routes.dart';

const Set<String> _loggedOutRoutes = {AppRoutes.login, AppRoutes.signup};

const Set<String> _authRoutes = {
  AppRoutes.login,
  AppRoutes.signup,
  AppRoutes.accountSetup,
  AppRoutes.splash,
};

/// Where [location] must go for [session], or null to stay:
/// - loading (no value yet) or failed: the splash, never a flash of `/login`;
/// - logged out: only login and signup;
/// - no username yet: account setup, from every route (Review Focus 3);
/// - signed in: auth routes lead home.
String? authRedirect(AsyncValue<Session?> session, String location) {
  if (!session.hasValue || session.hasError) {
    return location == AppRoutes.splash ? null : AppRoutes.splash;
  }
  final current = session.value;
  if (current == null) {
    return _loggedOutRoutes.contains(location) ? null : AppRoutes.login;
  }
  if (current.needsAccountSetup) {
    return location == AppRoutes.accountSetup ? null : AppRoutes.accountSetup;
  }
  return _authRoutes.contains(location) ? AppRoutes.home : null;
}

/// The app router. Re-runs [authRedirect] whenever [sessionProvider]
/// changes. The signed-in app is the [AppShell] over the registered
/// [appTabs]; pushed screens (the channel, settings) and auth screens sit outside it.
final appRouterProvider = Provider<GoRouter>((ref) {
  final session = ValueNotifier<AsyncValue<Session?>>(
    ref.read(sessionProvider),
  );
  ref.listen(sessionProvider, (_, next) => session.value = next);

  final router = GoRouter(
    initialLocation: AppRoutes.home,
    refreshListenable: session,
    redirect: (context, state) =>
        authRedirect(session.value, state.matchedLocation),
    routes: [
      AppShell.route(appTabs),
      // Above the shell: the player gets the whole screen, no tab bar.
      GoRoute(
        path: AppRoutes.channelPattern,
        builder: (context, state) =>
            ChannelScreen(userId: state.pathParameters['userId']!),
      ),
      GoRoute(
        path: AppRoutes.giftsPattern,
        builder: (context, state) =>
            GiftsScreen(userId: state.pathParameters['userId']!),
      ),
      GoRoute(
        path: AppRoutes.wallet,
        builder: (context, state) => const WalletScreen(),
      ),
      GoRoute(
        path: AppRoutes.shop,
        builder: (context, state) => const ShopScreen(),
      ),
      GoRoute(
        path: AppRoutes.vodPattern,
        builder: (context, state) => VodScreen(
          userId: state.pathParameters['userId']!,
          vodId: state.pathParameters['vodId']!,
        ),
      ),
      // Conversations also sit above the shell: the composer needs the
      // whole screen. `new` must come before the id pattern.
      GoRoute(
        path: AppRoutes.newConversation,
        builder: (context, state) => const NewConversationScreen(),
      ),
      GoRoute(
        path: AppRoutes.conversationPattern,
        builder: (context, state) => ConversationScreen(
          conversationId: state.pathParameters['conversationId']!,
        ),
      ),
      GoRoute(
        path: AppRoutes.settings,
        builder: (context, state) => const SettingsScreen(),
      ),
      GoRoute(
        path: AppRoutes.settingsProfile,
        builder: (context, state) => const ProfileSettingsScreen(),
      ),
      GoRoute(
        path: AppRoutes.settingsSecurity,
        builder: (context, state) => const SecurityScreen(),
      ),
      GoRoute(
        path: AppRoutes.login,
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: AppRoutes.signup,
        builder: (context, state) => const SignupScreen(),
      ),
      GoRoute(
        path: AppRoutes.accountSetup,
        builder: (context, state) => const AccountSetupScreen(),
      ),
      GoRoute(
        path: AppRoutes.splash,
        builder: (context, state) => const SplashScreen(),
      ),
    ],
  );
  ref.onDispose(() {
    router.dispose();
    session.dispose();
  });
  return router;
});
