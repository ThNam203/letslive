import 'package:flutter/material.dart';
import 'package:letslive_mobile/core/config/app_config.dart';

class LetsLiveApp extends StatelessWidget {
  const LetsLiveApp({required this.config, super.key});

  final AppConfig config;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'LetsLive',
      theme: ThemeData(colorSchemeSeed: Colors.deepPurple, useMaterial3: true),
      home: Scaffold(
        body: Center(child: Text('LetsLive (${config.flavor.name})')),
      ),
    );
  }
}

/// Shown when startup configuration is missing or invalid.
class ConfigErrorApp extends StatelessWidget {
  const ConfigErrorApp({required this.error, super.key});

  final ConfigException error;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text(
                    'App configuration error',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 12),
                  Text(error.message, textAlign: TextAlign.center),
                  const SizedBox(height: 12),
                  const Text(
                    'Run with --dart-define-from-file=env/<flavor>.json '
                    '(see .env.example).',
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
