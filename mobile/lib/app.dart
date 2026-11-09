import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/push/push_providers.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/core/theme/app_theme.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';

/// The app: theme, locale and the router (auth gate included).
class LetsLiveApp extends ConsumerWidget {
  const LetsLiveApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // Device registration and push taps live as long as the app.
    ref.watch(pushBridgeProvider);
    return MaterialApp.router(
      routerConfig: ref.watch(appRouterProvider),
      onGenerateTitle: (context) =>
          AppLocalizations.of(context).common_app_title,
      theme: lightTheme,
      darkTheme: darkTheme,
      themeMode: ref.watch(themeModeProvider),
      locale: ref.watch(localeProvider),
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
    );
  }
}

/// Shown when startup configuration is missing or invalid.
class ConfigErrorApp extends StatelessWidget {
  const ConfigErrorApp({required this.error, super.key});

  final ConfigException error;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text(
                    'App configuration error',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 12),
                  Text(error.message, textAlign: TextAlign.center),
                  const SizedBox(height: 12),
                  const Text(
                    'Run with --dart-define-from-file=env/<flavor>.json '
                    '(see .env.example).',
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
