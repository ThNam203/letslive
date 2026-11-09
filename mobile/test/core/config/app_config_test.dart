import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/app_config.dart';

void main() {
  const webExampleEnv = <String, String>{
    'BACKEND_PROTOCOL': 'http',
    'BACKEND_WS_PROTOCOL': 'ws',
    'BACKEND_HOST': 'localhost',
    'BACKEND_PORT': '8000',
    'FLAVOR': 'dev',
    'TURNSTILE_DEV_TOKEN': 'dev-token',
  };

  const firebaseEnv = <String, String>{
    'FIREBASE_PROJECT_ID': 'letslive-test',
    'FIREBASE_API_KEY': 'api-key',
    'FIREBASE_MESSAGING_SENDER_ID': '1234',
    'FIREBASE_ANDROID_APP_ID': '1:1234:android:abc',
    'FIREBASE_IOS_APP_ID': '1:1234:ios:abc',
  };

  final releaseEnv = <String, String>{
    ...webExampleEnv,
    ...firebaseEnv,
    'FLAVOR': 'prod',
    'GOOGLE_SERVER_CLIENT_ID': 'web.apps.googleusercontent.com',
    'GOOGLE_IOS_CLIENT_ID': 'ios.apps.googleusercontent.com',
    'WEB_BASE_URL': 'https://letslive.example',
  };

  Matcher configErrorNaming(String key) => throwsA(
    isA<ConfigException>().having((e) => e.message, 'message', contains(key)),
  );

  group('AppConfig.fromEnv', () {
    test('builds http and ws urls from env', () {
      final config = AppConfig.fromEnv(webExampleEnv);

      expect(config.apiBaseUrl, 'http://localhost:8000');
      expect(config.wsBaseUrl, 'ws://localhost:8000/realtime');
      expect(config.flavor, Flavor.dev);
    });

    test('throws when host is missing', () {
      final env = {...webExampleEnv}..remove('BACKEND_HOST');

      expect(
        () => AppConfig.fromEnv(env),
        throwsA(
          isA<ConfigException>().having(
            (e) => e.message,
            'message',
            contains('BACKEND_HOST'),
          ),
        ),
      );
    });

    test('throws when a required value is blank', () {
      final env = {...webExampleEnv, 'BACKEND_PORT': '  '};

      expect(() => AppConfig.fromEnv(env), throwsA(isA<ConfigException>()));
    });

    test('throws on unknown flavor', () {
      final env = {...webExampleEnv, 'FLAVOR': 'qa'};

      expect(() => AppConfig.fromEnv(env), throwsA(isA<ConfigException>()));
    });

    test('throws on non-numeric port', () {
      final env = {...webExampleEnv, 'BACKEND_PORT': 'abc'};

      expect(() => AppConfig.fromEnv(env), throwsA(isA<ConfigException>()));
    });

    test('parses staging and prod flavors', () {
      expect(
        AppConfig.fromEnv({...releaseEnv, 'FLAVOR': 'staging'}).flavor,
        Flavor.staging,
      );
      expect(AppConfig.fromEnv(releaseEnv).flavor, Flavor.prod);
    });
  });

  group('AppConfig.fromEnv auth settings', () {
    test('dev without Firebase uses the Turnstile dev token', () {
      final config = AppConfig.fromEnv(webExampleEnv);

      expect(config.firebase, isNull);
      expect(config.turnstileDevToken, 'dev-token');
      expect(config.googleSignIn, isNull);
    });

    test('dev without Firebase requires TURNSTILE_DEV_TOKEN', () {
      final env = {...webExampleEnv}..remove('TURNSTILE_DEV_TOKEN');

      expect(
        () => AppConfig.fromEnv(env),
        configErrorNaming('TURNSTILE_DEV_TOKEN'),
      );
    });

    test('dev with Firebase needs no Turnstile token', () {
      final env = {...webExampleEnv, ...firebaseEnv}
        ..remove('TURNSTILE_DEV_TOKEN');

      final config = AppConfig.fromEnv(env, platform: TargetPlatform.android);

      expect(config.firebase?.projectId, 'letslive-test');
      expect(config.firebase?.appId, '1:1234:android:abc');
      expect(config.turnstileDevToken, isNull);
    });

    test('Firebase uses the app id of the running platform', () {
      final config = AppConfig.fromEnv(
        releaseEnv,
        platform: TargetPlatform.iOS,
      );

      expect(config.firebase?.appId, '1:1234:ios:abc');
    });

    test('Firebase config names a missing companion key', () {
      final env = {...releaseEnv}..remove('FIREBASE_API_KEY');

      expect(
        () => AppConfig.fromEnv(env),
        configErrorNaming('FIREBASE_API_KEY'),
      );
    });

    test('prod requires Firebase', () {
      final env = {...releaseEnv}..remove('FIREBASE_PROJECT_ID');

      expect(
        () => AppConfig.fromEnv(env),
        configErrorNaming('FIREBASE_PROJECT_ID'),
      );
    });

    test('staging requires the Google server client id', () {
      final env = {...releaseEnv, 'FLAVOR': 'staging'}
        ..remove('GOOGLE_SERVER_CLIENT_ID');

      expect(
        () => AppConfig.fromEnv(env),
        configErrorNaming('GOOGLE_SERVER_CLIENT_ID'),
      );
    });

    test('staging and prod require the web base url', () {
      for (final flavor in ['staging', 'prod']) {
        final env = {...releaseEnv, 'FLAVOR': flavor}..remove('WEB_BASE_URL');

        expect(() => AppConfig.fromEnv(env), configErrorNaming('WEB_BASE_URL'));
      }
    });

    test('the web base url loses a trailing slash; dev may leave it out', () {
      final withSlash = AppConfig.fromEnv({
        ...releaseEnv,
        'WEB_BASE_URL': 'https://letslive.example//',
      });
      expect(withSlash.webBaseUrl, 'https://letslive.example');
      expect(AppConfig.fromEnv(webExampleEnv).webBaseUrl, isNull);
    });

    test('Google sign-in on iOS requires the iOS client id', () {
      final env = {...releaseEnv}..remove('GOOGLE_IOS_CLIENT_ID');

      expect(
        () => AppConfig.fromEnv(env, platform: TargetPlatform.iOS),
        configErrorNaming('GOOGLE_IOS_CLIENT_ID'),
      );
      final android = AppConfig.fromEnv(env, platform: TargetPlatform.android);
      expect(
        android.googleSignIn?.serverClientId,
        'web.apps.googleusercontent.com',
      );
      expect(android.googleSignIn?.iosClientId, isNull);
    });

    test('prod ignores the Turnstile dev token', () {
      final config = AppConfig.fromEnv(releaseEnv);

      expect(config.turnstileDevToken, isNull);
    });
  });
}
