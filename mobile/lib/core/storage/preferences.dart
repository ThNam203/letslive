import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// The app-wide [SharedPreferences] instance.
///
/// `main` loads it before `runApp` and overrides this provider, so reading it
/// without an override is a wiring bug and fails loudly.
final sharedPreferencesProvider = Provider<SharedPreferences>(
  (ref) => throw StateError(
    'sharedPreferencesProvider must be overridden with a loaded instance.',
  ),
);
