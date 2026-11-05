import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';

import 'profile_test_app.dart';

Finder _key(String key) => find.byKey(ValueKey(key));

/// Theme and language pickers of the settings hub (web `/settings/profile`
/// theme and language lists).
void main() {
  testWidgets('choosing Dark switches the app theme and is remembered', (
    tester,
  ) async {
    final app = await pumpRoute(
      tester,
      FakeProfileServer(),
      AppRoutes.settings,
    );

    await tester.tap(_key('settings-theme'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Dark'));
    await tester.pumpAndSettle();

    final material = tester.widget<MaterialApp>(find.byType(MaterialApp));
    expect(material.themeMode, ThemeMode.dark);
    final prefs = app.container.read(sharedPreferencesProvider);
    expect(prefs.getString(themeModePreferenceKey), 'dark');
    // The tile shows the choice.
    expect(
      find.descendant(of: _key('settings-theme'), matching: find.text('Dark')),
      findsOneWidget,
    );
  });

  testWidgets('choosing Tiếng Việt switches the UI language, is remembered '
      'and saved on the profile like web', (tester) async {
    final server = FakeProfileServer();
    final app = await pumpRoute(tester, server, AppRoutes.settings);
    expect(find.text('Settings'), findsOneWidget);

    await tester.tap(_key('settings-language'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Tiếng Việt'));
    await tester.pumpAndSettle();

    expect(find.text('Settings'), findsNothing);
    expect(find.text('Cài đặt'), findsOneWidget);
    final prefs = app.container.read(sharedPreferencesProvider);
    expect(prefs.getString(localePreferenceKey), 'vi');
    expect(server.profileUpdates.last, {'locale': 'vi-VN'});
  });

  testWidgets('the pickers fit 320px at text scale 2', (tester) async {
    await pumpRoute(
      tester,
      FakeProfileServer(),
      AppRoutes.settings,
      width: 320,
      height: 568,
      textScale: 2,
    );
    expect(tester.takeException(), isNull);

    await tester.tap(_key('settings-theme'));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.tap(find.text('System').last);
    await tester.pumpAndSettle();

    await tester.tap(_key('settings-language'));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
