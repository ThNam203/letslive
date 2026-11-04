import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/theme/app_theme.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';

class LetsLiveApp extends ConsumerWidget {
  const LetsLiveApp({required this.config, super.key});

  final AppConfig config;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp(
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
      home: Scaffold(
        body: Center(child: Text('LetsLive (${config.flavor.name})')),
      ),
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
