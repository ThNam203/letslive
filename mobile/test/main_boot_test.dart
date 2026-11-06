import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/main.dart';

const _firebase = FirebaseConfig(
  projectId: 'letslive-dev',
  apiKey: 'AIza-secret-key',
  messagingSenderId: '123456789',
  appId: '1:123456789:android:abc',
);

void main() {
  test('a successful Firebase setup boots normally', () async {
    expect(await guardFirebaseSetup(_firebase, () async {}), isNull);
  });

  test('a Firebase error becomes a config error without secrets', () async {
    final error = await guardFirebaseSetup(
      _firebase,
      () async => throw FirebaseException(
        plugin: 'core',
        code: 'invalid-api-key',
        message: 'API key AIza-secret-key is not valid',
      ),
    );

    expect(error, isA<ConfigException>());
    expect(error!.message, contains('letslive-dev'));
    expect(error.message, contains('invalid-api-key'));
    expect(error.message, isNot(contains('AIza-secret-key')));
    expect(error.message, isNot(contains('123456789')));
  });

  test('a platform error (e.g. App Check activate) is caught too', () async {
    final error = await guardFirebaseSetup(
      _firebase,
      () async => throw PlatformException(code: 'app-check'),
    );

    expect(error?.message, contains('PlatformException'));
  });
}
