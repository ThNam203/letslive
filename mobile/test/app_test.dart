import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';
import 'package:letslive_mobile/features/home/presentation/home_screen.dart';

import 'features/auth/auth_test_app.dart';

void main() {
  testWidgets('ConfigErrorApp shows the config error message', (tester) async {
    await tester.pumpWidget(
      const ConfigErrorApp(
        error: ConfigException('Missing required config value BACKEND_HOST'),
      ),
    );

    expect(find.text('App configuration error'), findsOneWidget);
    expect(find.textContaining('BACKEND_HOST'), findsOneWidget);
  });

  testWidgets('LetsLiveApp renders with a valid config', (tester) async {
    await pumpTestApp(tester, signedIn: true);

    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  testWidgets('LetsLiveApp applies the persisted locale and theme', (
    tester,
  ) async {
    await pumpTestApp(
      tester,
      signedIn: true,
      preferences: {localePreferenceKey: 'vi', themeModePreferenceKey: 'dark'},
    );

    final app = tester.widget<MaterialApp>(find.byType(MaterialApp));
    expect(app.locale, const Locale('vi'));
    expect(app.themeMode, ThemeMode.dark);
    final context = tester.element(find.byType(HomeScreen));
    expect(AppLocalizations.of(context).localeName, 'vi');
    expect(Theme.of(context).brightness, Brightness.dark);
  });
}
