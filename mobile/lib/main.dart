import 'dart:developer' as developer;

import 'package:firebase_app_check/firebase_app_check.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/config/env.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Boot: parse the env (a config error gets its own screen), set up Firebase
/// App Check when configured, start the realtime socket once, run the app.
Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final AppConfig config;
  try {
    config = AppConfig.fromEnv(compileTimeEnv);
  } on ConfigException catch (e) {
    runApp(ConfigErrorApp(error: e));
    return;
  }

  final firebase = config.firebase;
  if (firebase != null) {
    final error = await guardFirebaseSetup(
      firebase,
      () => _initFirebase(firebase, debug: config.flavor == Flavor.dev),
    );
    if (error != null) {
      runApp(ConfigErrorApp(error: error));
      return;
    }
  }
  final prefs = await SharedPreferences.getInstance();
  final container = ProviderContainer(
    overrides: [
      appConfigProvider.overrideWithValue(config),
      sharedPreferencesProvider.overrideWithValue(prefs),
    ],
  );
  container.read(realtimeManagerProvider).start();
  runApp(
    UncontrolledProviderScope(container: container, child: const LetsLiveApp()),
  );
}

/// Runs the Firebase [setUp]. A failure (bad FIREBASE_* values, App Check not
/// set up) is logged and returned as a [ConfigException] for the config
/// error screen instead of crashing the boot. The message names only the
/// project id and the error type/code, never the key values.
@visibleForTesting
Future<ConfigException?> guardFirebaseSetup(
  FirebaseConfig firebase,
  Future<void> Function() setUp,
) async {
  try {
    await setUp();
    return null;
  } on Exception catch (e, stack) {
    developer.log(
      'Firebase setup failed',
      name: 'boot',
      error: e,
      stackTrace: stack,
    );
    final reason = e is FirebaseException ? e.code : '${e.runtimeType}';
    return ConfigException(
      'Firebase setup failed for project "${firebase.projectId}" ($reason). '
      'Check the FIREBASE_* values and the App Check setup.',
    );
  }
}

/// Firebase from env values (no google-services files). Dev uses the App
/// Check debug provider: register the debug token it logs in the Firebase
/// console.
Future<void> _initFirebase(
  FirebaseConfig firebase, {
  required bool debug,
}) async {
  await Firebase.initializeApp(
    options: FirebaseOptions(
      apiKey: firebase.apiKey,
      appId: firebase.appId,
      messagingSenderId: firebase.messagingSenderId,
      projectId: firebase.projectId,
    ),
  );
  await FirebaseAppCheck.instance.activate(
    providerAndroid: debug
        ? const AndroidDebugProvider()
        : const AndroidPlayIntegrityProvider(),
    providerApple: debug
        ? const AppleDebugProvider()
        : const AppleAppAttestWithDeviceCheckFallbackProvider(),
  );
}
