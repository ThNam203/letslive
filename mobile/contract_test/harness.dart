import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/config/env.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

/// Cloudflare Turnstile token sent on `/auth/verify-email`. The dev stack's
/// test secret (`1x0000000000000000000000000000000AA`) accepts any non-empty
/// token. The User-Agent is irrelevant to the captcha check.
const String _turnstileToken = String.fromEnvironment(
  'TURNSTILE_DEV_TOKEN',
  defaultValue: '',
);

/// Postgres container holding the auth database (`docker exec` target).
const String _dbContainer = String.fromEnvironment(
  'CONTRACT_DB_CONTAINER',
  defaultValue: 'll-mobile-db',
);
const String _dbUser = String.fromEnvironment(
  'CONTRACT_DB_USER',
  defaultValue: 'postgres',
);
const String _authDb = String.fromEnvironment(
  'CONTRACT_AUTH_DB',
  defaultValue: 'letslive_auth',
);

/// Only identifies the harness in server logs. The backend no longer treats any
/// User-Agent specially; the captcha is satisfied by the Turnstile token.
const String _userAgent = 'letslive-contract-tests/1.0';

const Duration _otpWait = Duration(seconds: 10);
const Duration _userReadyWait = Duration(seconds: 20);
const Duration _pollInterval = Duration(milliseconds: 300);
const String _testPassword = 'Contract-Pass1!';

/// Thrown when the contract environment is missing or the stack misbehaves.
class ContractSetupException implements Exception {
  const ContractSetupException(this.message);

  final String message;

  @override
  String toString() => 'ContractSetupException: $message';
}

/// A signed-up test user and a client authenticated as that user.
class ContractUser {
  const ContractUser({
    required this.email,
    required this.username,
    required this.password,
    required this.client,
    required this.cookieJar,
  });

  final String email;
  final String username;
  final String password;
  final ApiClient client;

  /// The session cookies of [client]; a `RealtimeManager` needs the same jar
  /// to open its socket as this user.
  final CookieJar cookieJar;
}

/// Two users created through the real signup path.
class ContractContext {
  const ContractContext({
    required this.config,
    required this.userA,
    required this.userB,
  });

  final AppConfig config;
  final ContractUser userA;
  final ContractUser userB;

  Future<void> dispose() async {
    await userA.client.dispose();
    await userB.client.dispose();
  }
}

/// Creates two fresh users (unique emails per run) via
/// `POST /auth/verify-email` + `POST /auth/signup` against the dev stack and
/// returns authenticated clients (cookie jar per user).
///
/// The OTP is read from the dev Postgres with `docker exec ... psql`: the auth
/// service inserts the OTP before it tries to email it, so it is present even
/// though the dev stack cannot send mail. Needs a running stack, the usual
/// `BACKEND_*` defines and `TURNSTILE_DEV_TOKEN`.
Future<ContractContext> startContractContext() async {
  if (_turnstileToken.isEmpty) {
    throw const ContractSetupException(
      'TURNSTILE_DEV_TOKEN is not set (add it to the dart-define file).',
    );
  }
  final AppConfig config;
  try {
    config = AppConfig.fromEnv(compileTimeEnv);
  } on ConfigException catch (e) {
    throw ContractSetupException(e.message);
  }

  final runId = _runId();
  final userA = await _signUp(config, 'ca$runId');
  try {
    final userB = await _signUp(config, 'cb$runId');
    return ContractContext(config: config, userA: userA, userB: userB);
  } on Object {
    await userA.client.dispose();
    rethrow;
  }
}

String _runId() {
  final millis = DateTime.now().millisecondsSinceEpoch.toRadixString(36);
  final noise = Random.secure().nextInt(36 * 36).toRadixString(36);
  return '$millis$noise';
}

Future<ContractUser> _signUp(AppConfig config, String name) async {
  final email = '$name@contract.test';
  final username = 'user_$name';
  final jar = CookieJar();
  final client = ApiClient(
    baseUrl: config.apiBaseUrl,
    cookieJar: jar,
    defaultHeaders: const {'User-Agent': _userAgent},
  );
  try {
    // Delivery fails on the dev stack (no SMTP credentials), so the response
    // is not checked; the OTP row existing is what matters.
    final verify = await client.post<Object?>(
      ApiEndpoints.verifyEmail,
      body: {'email': email, 'turnstileToken': _turnstileToken},
      parse: (data) => data,
    );
    final otp = await readSignupOtp(email, verifyKey: verify.key);

    final signup = await client.post<Object?>(
      ApiEndpoints.signup,
      body: {
        'email': email,
        'username': username,
        'password': _testPassword,
        'otpCode': otp,
        'turnstileToken': _turnstileToken,
      },
      parse: (data) => data,
    );
    if (!signup.success) {
      throw ContractSetupException(
        'signup failed for $email: ${signup.statusCode} ${signup.key} '
        '${signup.message}',
      );
    }
    await _waitUntilUserExists(client, email);
    return ContractUser(
      email: email,
      username: username,
      password: _testPassword,
      client: client,
      cookieJar: jar,
    );
  } on Object {
    await client.dispose();
    rethrow;
  }
}

/// The user service creates its row asynchronously after signup.
Future<void> _waitUntilUserExists(ApiClient client, String email) async {
  final deadline = DateTime.now().add(_userReadyWait);
  var last = 'no attempt';
  while (DateTime.now().isBefore(deadline)) {
    final res = await client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    if (res.success) return;
    last = '${res.statusCode} ${res.key} ${res.message}';
    await Future<void>.delayed(_pollInterval);
  }
  throw ContractSetupException(
    'GET /user/me never succeeded for $email: $last',
  );
}

final RegExp _safeEmail = RegExp(r'^[a-z0-9]+@[a-z0-9.]+$');

/// The latest signup code stored for [email] (see [startContractContext] for
/// why it is read from the database). [verifyKey] only enriches the error.
Future<String> readSignupOtp(String email, {String verifyKey = ''}) async {
  // The email is interpolated into SQL; callers generate it, never external.
  if (!_safeEmail.hasMatch(email)) {
    throw ContractSetupException('refusing unsafe email "$email"');
  }
  final sql =
      "SELECT code FROM sign_up_otps WHERE email = '$email' "
      'ORDER BY created_at DESC LIMIT 1';
  final deadline = DateTime.now().add(_otpWait);
  while (true) {
    final code = await runSql(_authDb, sql);
    if (code.isNotEmpty) return code;
    if (DateTime.now().isAfter(deadline)) {
      throw ContractSetupException(
        'no OTP row for $email after $_otpWait '
        '(verify-email key: "$verifyKey")',
      );
    }
    await Future<void>.delayed(_pollInterval);
  }
}

/// Runs [sql] in database [database] of the dev Postgres container
/// (`docker exec ... psql`) and returns the trimmed output. Contract tests use
/// it for fixtures the public API cannot create (e.g. a live stream needs a
/// real RTMP publisher).
Future<String> runSql(String database, String sql) async {
  final ProcessResult result;
  try {
    result = await Process.run('docker', [
      'exec',
      _dbContainer,
      'psql',
      '-U',
      _dbUser,
      '-d',
      database,
      '-tA',
      '-c',
      sql,
    ], stdoutEncoding: utf8);
  } on ProcessException catch (e) {
    throw ContractSetupException('cannot run docker: ${e.message}');
  }
  if (result.exitCode != 0) {
    throw ContractSetupException(
      'psql in container "$_dbContainer" failed: ${result.stderr}',
    );
  }
  return (result.stdout as String).trim();
}
