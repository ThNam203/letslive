import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';
import 'package:letslive_mobile/core/theme/theme_mode_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

Future<ProviderContainer> makeContainer(
  Map<String, Object> stored, {
  Locale device = const Locale('en', 'US'),
}) async {
  SharedPreferences.setMockInitialValues(stored);
  final prefs = await SharedPreferences.getInstance();
  final container = ProviderContainer(
    overrides: [
      sharedPreferencesProvider.overrideWithValue(prefs),
      deviceLocaleProvider.overrideWithValue(device),
    ],
  );
  addTearDown(container.dispose);
  return container;
}

void main() {
  group('themeModeProvider', () {
    test('defaults to system', () async {
      final container = await makeContainer({});

      expect(container.read(themeModeProvider), ThemeMode.system);
    });

    test('restores the persisted mode', () async {
      final container = await makeContainer({themeModePreferenceKey: 'dark'});

      expect(container.read(themeModeProvider), ThemeMode.dark);
    });

    test('ignores an unknown persisted value', () async {
      final container = await makeContainer({themeModePreferenceKey: 'neon'});

      expect(container.read(themeModeProvider), ThemeMode.system);
    });

    test('persists changes', () async {
      final container = await makeContainer({});

      await container
          .read(themeModeProvider.notifier)
          .setThemeMode(ThemeMode.light);

      expect(container.read(themeModeProvider), ThemeMode.light);
      final prefs = container.read(sharedPreferencesProvider);
      expect(prefs.getString(themeModePreferenceKey), 'light');
    });
  });

  group('localeProvider', () {
    test('follows a supported device language', () async {
      final container = await makeContainer(
        {},
        device: const Locale('vi', 'VN'),
      );

      expect(container.read(localeProvider), const Locale('vi'));
    });

    test('falls back to en for an unsupported device language', () async {
      final container = await makeContainer(
        {},
        device: const Locale('fr', 'FR'),
      );

      expect(container.read(localeProvider), const Locale('en'));
    });

    test('persisted choice wins over the device language', () async {
      final container = await makeContainer({
        localePreferenceKey: 'en',
      }, device: const Locale('vi', 'VN'));

      expect(container.read(localeProvider), const Locale('en'));
    });

    test('persists changes', () async {
      final container = await makeContainer({});

      await container
          .read(localeProvider.notifier)
          .setLocale(const Locale('vi'));

      expect(container.read(localeProvider), const Locale('vi'));
      final prefs = container.read(sharedPreferencesProvider);
      expect(prefs.getString(localePreferenceKey), 'vi');
    });
  });
}
