import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/app_config.dart';

void main() {
  const webExampleEnv = <String, String>{
    'BACKEND_PROTOCOL': 'http',
    'BACKEND_WS_PROTOCOL': 'ws',
    'BACKEND_HOST': 'localhost',
    'BACKEND_PORT': '8000',
    'FLAVOR': 'dev',
  };

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
        AppConfig.fromEnv({...webExampleEnv, 'FLAVOR': 'staging'}).flavor,
        Flavor.staging,
      );
      expect(
        AppConfig.fromEnv({...webExampleEnv, 'FLAVOR': 'prod'}).flavor,
        Flavor.prod,
      );
    });
  });
}
