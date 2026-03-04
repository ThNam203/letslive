import 'env.dart';

abstract final class AppConfig {
  static String get apiUrl =>
      '${Env.backendProtocol}://${Env.backendIpAddress}:${Env.backendPort}';

  static String get wsUrl =>
      '${Env.backendWsProtocol}://${Env.backendIpAddress}:${Env.backendPort}/ws';

  static String get dmWsUrl =>
      '${Env.backendWsProtocol}://${Env.backendIpAddress}:${Env.backendPort}/dm-ws';

  static const requestTimeout = Duration(seconds: 15);
}
