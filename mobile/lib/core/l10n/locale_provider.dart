import 'dart:ui';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';

const localePreferenceKey = 'locale';
const fallbackLocale = Locale('en');
const supportedAppLocales = [Locale('en'), Locale('vi')];

/// The device's preferred locale. A provider so tests can override it.
final deviceLocaleProvider = Provider<Locale>(
  (ref) => PlatformDispatcher.instance.locale,
);

Locale _resolve(String? languageCode) => supportedAppLocales.firstWhere(
  (locale) => locale.languageCode == languageCode,
  orElse: () => fallbackLocale,
);

/// User-selected app language, persisted in shared_preferences. With nothing
/// stored it follows the device language, falling back to English.
class LocaleNotifier extends Notifier<Locale> {
  @override
  Locale build() {
    final stored = ref
        .watch(sharedPreferencesProvider)
        .getString(localePreferenceKey);
    if (stored != null) return _resolve(stored);
    return _resolve(ref.watch(deviceLocaleProvider).languageCode);
  }

  Future<void> setLocale(Locale locale) async {
    final resolved = _resolve(locale.languageCode);
    state = resolved;
    await ref
        .read(sharedPreferencesProvider)
        .setString(localePreferenceKey, resolved.languageCode);
  }
}

final localeProvider = NotifierProvider<LocaleNotifier, Locale>(
  LocaleNotifier.new,
);
