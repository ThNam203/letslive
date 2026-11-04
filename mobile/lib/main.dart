import 'package:flutter/material.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/config/env.dart';

void main() {
  try {
    final config = AppConfig.fromEnv(compileTimeEnv);
    runApp(LetsLiveApp(config: config));
  } on ConfigException catch (e) {
    runApp(ConfigErrorApp(error: e));
  }
}
