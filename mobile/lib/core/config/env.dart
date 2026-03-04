import 'package:flutter_dotenv/flutter_dotenv.dart';

abstract final class Env {
  static String get backendProtocol => dotenv.env['BACKEND_PROTOCOL'] ?? 'http';
  static String get backendIpAddress => dotenv.env['BACKEND_IP_ADDRESS'] ?? 'localhost';
  static String get backendPort => dotenv.env['BACKEND_PORT'] ?? '8000';
  static String get backendWsProtocol => dotenv.env['BACKEND_WS_PROTOCOL'] ?? 'ws';
  static String get environment => dotenv.env['ENVIRONMENT'] ?? 'development';

  static bool get isDevelopment => environment == 'development';
  static bool get isProduction => environment == 'production';
}
