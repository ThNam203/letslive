import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/domain/session.dart';
import 'package:letslive_mobile/features/auth/presentation/account_setup_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/signup_screen.dart';
import 'package:letslive_mobile/features/auth/presentation/splash_screen.dart';

/// Route paths. Auth routes mirror web `app/(auth)`.
abstract final class AppRoutes {
  static const String home = '/';
  static const String login = '/login';
  static const String signup = '/signup';
  static const String accountSetup = '/account-setup';

  /// Shown while the session loads, or with a retry if loading failed.
  static const String splash = '/splash';
}

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
/// changes. Task 11 adds the shell and feature routes.
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
      GoRoute(
        path: AppRoutes.home,
        // The scaffold's root body until Task 11 replaces it with the shell.
        builder: (context, state) => const _RootBody(),
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

class _RootBody extends ConsumerWidget {
  const _RootBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    body: Center(
      child: Text('LetsLive (${ref.watch(appConfigProvider).flavor.name})'),
    ),
  );
}
