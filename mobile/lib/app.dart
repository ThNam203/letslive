import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import 'package:letslive/l10n/app_localizations.dart';

import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';
import 'providers.dart';

class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final themeMode = ref.watch(themeModeProvider);
    final locale = ref.watch(localeProvider);

    // Resolve effective brightness.
    final brightness = switch (themeMode) {
      AppThemeMode.light => Brightness.light,
      AppThemeMode.dark => Brightness.dark,
      AppThemeMode.system => MediaQuery.platformBrightnessOf(context),
    };
    final theme =
        brightness == Brightness.dark ? AppTheme.dark : AppTheme.light;

    return MaterialApp.router(
      title: "Let's Live",
      debugShowCheckedModeBanner: false,

      // Forui theme → approximate Material theme for compatibility
      theme: theme.toApproximateMaterialTheme(),

      // Locale
      locale: locale,

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
        child: FToaster(child: child!),
      ),

      // Router
      routerConfig: appRouter,
    );
  }
}
