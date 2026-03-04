import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:forui/forui.dart';

import 'package:letslive/l10n/app_localizations.dart';

import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';

class App extends StatelessWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context) {
    // Use system brightness to pick light/dark theme.
    final brightness = MediaQuery.platformBrightnessOf(context);
    final theme =
        brightness == Brightness.dark ? AppTheme.dark : AppTheme.light;

    return MaterialApp.router(
      title: "Let's Live",
      debugShowCheckedModeBanner: false,

      // Forui theme → approximate Material theme for compatibility
      theme: theme.toApproximateMaterialTheme(),

      // i18n
      localizationsDelegates: const [
        AppLocalizations.delegate,
        ...FLocalizations.localizationsDelegates,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,

      // Wrap with FTheme + FToaster for Forui widgets
      builder: (_, child) => FTheme(
        data: theme,
        child: child!,
      ),

      // Router
      routerConfig: appRouter,
    );
  }
}
