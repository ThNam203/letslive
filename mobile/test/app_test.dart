import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

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

  Future<void> pumpApp(WidgetTester tester, Map<String, Object> stored) async {
    SharedPreferences.setMockInitialValues(stored);
    final prefs = await SharedPreferences.getInstance();
    final config = AppConfig.fromEnv(const {
      'BACKEND_PROTOCOL': 'http',
      'BACKEND_WS_PROTOCOL': 'ws',
      'BACKEND_HOST': 'localhost',
      'BACKEND_PORT': '8000',
      'FLAVOR': 'dev',
    });
    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPreferencesProvider.overrideWithValue(prefs)],
        child: LetsLiveApp(config: config),
      ),
    );
    await tester.pump();
  }

  testWidgets('LetsLiveApp renders with a valid config', (tester) async {
    await pumpApp(tester, {});

    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.text('LetsLive (dev)'), findsOneWidget);
  });

  testWidgets('LetsLiveApp applies the persisted locale and theme', (
    tester,
  ) async {
    await pumpApp(tester, {
      localePreferenceKey: 'vi',
      themeModePreferenceKey: 'dark',
    });

    final app = tester.widget<MaterialApp>(find.byType(MaterialApp));
    expect(app.locale, const Locale('vi'));
    expect(app.themeMode, ThemeMode.dark);
    final context = tester.element(find.text('LetsLive (dev)'));
    expect(AppLocalizations.of(context).localeName, 'vi');
    expect(Theme.of(context).brightness, Brightness.dark);
  });
}
