import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/config/env.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';
import 'package:shared_preferences/shared_preferences.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    final config = AppConfig.fromEnv(compileTimeEnv);
    final prefs = await SharedPreferences.getInstance();
    runApp(
      ProviderScope(
        overrides: [sharedPreferencesProvider.overrideWithValue(prefs)],
        child: LetsLiveApp(config: config),
      ),
    );
  } on ConfigException catch (e) {
    runApp(ConfigErrorApp(error: e));
  }
}
