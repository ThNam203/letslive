import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';

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

  testWidgets('LetsLiveApp renders with a valid config', (tester) async {
    final config = AppConfig.fromEnv(const {
      'BACKEND_PROTOCOL': 'http',
      'BACKEND_WS_PROTOCOL': 'ws',
      'BACKEND_HOST': 'localhost',
      'BACKEND_PORT': '8000',
      'FLAVOR': 'dev',
    });

    await tester.pumpWidget(LetsLiveApp(config: config));

    expect(find.byType(MaterialApp), findsOneWidget);
    expect(find.text('LetsLive (dev)'), findsOneWidget);
  });
}
