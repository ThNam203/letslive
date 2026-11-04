import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/app.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/storage/preferences.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';
import 'package:letslive_mobile/features/auth/data/google_sign_in_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../core/network/fake_adapter.dart';

const testBaseUrl = 'http://localhost:8000';
const testPassword = 'Secret-1!';
const testOtp = '123456';

/// Dev config as the widget tests run it.
final testConfig = AppConfig.fromEnv(const {
  'BACKEND_PROTOCOL': 'http',
  'BACKEND_WS_PROTOCOL': 'ws',
  'BACKEND_HOST': 'localhost',
  'BACKEND_PORT': '8000',
  'FLAVOR': 'dev',
  'TURNSTILE_DEV_TOKEN': 'dev-token',
});

ResponseBody _error(int status, String key, String message) => jsonBody({
  'requestId': 'req-err',
  'success': false,
  'code': 1,
  'key': key,
  'message': message,
}, status: status);

/// In-memory stand-in for the auth and `/user/me` endpoints. Keeps the
/// session in a real cookie (`ACCESS_TOKEN`), so the app's cookie jar and
/// refresh logic run unchanged.
class FakeAuthServer {
  FakeAuthServer({this.username = 'abcdef'});

  /// Username returned by `/user/me`; empty or null means account setup is
  /// pending.
  String? username;

  /// Status `/user/me` answers with for a valid session (e.g. 503).
  int meStatus = 200;

  /// When set, `/auth/login` waits for it before answering.
  Completer<void>? holdLogin;

  /// When set, `GET /user/me` waits for it before answering.
  Completer<void>? holdMe;

  /// When set, `PUT /user/me` (account setup) waits for it.
  Completer<void>? holdUpdateMe;

  /// When set, `/auth/verify-email` (send the signup code) waits for it.
  Completer<void>? holdVerifyEmail;

  /// When set, `/auth/signup` waits for it.
  Completer<void>? holdSignup;

  ResponseBody _me() => okEnvelope({
    'id': 'u-1',
    'username': username,
    'email': 'user@example.com',
    'status': 'normal',
    'authProvider': 'local',
    'createdAt': '2026-11-04T00:00:00Z',
    'followerCount': 0,
    'streamAPIKey': 'stream-key',
  });

  ResponseBody _signedIn() => jsonBody(
    {
      'requestId': 'req-ok',
      'success': true,
      'code': 0,
      'key': 'res_succ_login',
      'message': 'ok',
    },
    headers: {
      'set-cookie': ['ACCESS_TOKEN=access; Path=/'],
    },
  );

  bool _hasSession(RequestOptions o) =>
      '${o.headers['cookie'] ?? ''}'.contains('ACCESS_TOKEN=access');

  Map<String, Object?> _body(RequestOptions o) =>
      (o.data as Map<String, Object?>?) ?? const {};

  /// The [FakeHandler] for [FakeAdapter].
  Future<ResponseBody> handle(RequestOptions o) async {
    switch ((o.method, o.path)) {
      case ('GET', ApiEndpoints.userMe):
        await holdMe?.future;
        if (!_hasSession(o)) {
          return _error(401, 'res_err_unauthorized', 'Unauthorized');
        }
        if (meStatus != 200) {
          return _error(meStatus, 'res_err_internal_server', 'Server down');
        }
        return _me();
      case ('PUT', ApiEndpoints.userMe):
        await holdUpdateMe?.future;
        final wanted = _body(o)['username'] as String? ?? '';
        if (wanted == 'taken_name') {
          return _error(409, 'res_err_username_taken', 'Username is taken');
        }
        username = wanted;
        return _me();
      case ('POST', ApiEndpoints.login):
        await holdLogin?.future;
        if (_body(o)['password'] != testPassword) {
          return _error(
            401,
            'res_err_email_or_password_incorrect',
            'Username or password incorrect.',
          );
        }
        return _signedIn();
      case ('POST', ApiEndpoints.googleMobile):
        return _signedIn();
      case ('POST', ApiEndpoints.verifyEmail):
        await holdVerifyEmail?.future;
        return okEnvelope(null);
      case ('POST', ApiEndpoints.signup):
        await holdSignup?.future;
        if (_body(o)['otpCode'] != testOtp) {
          return _error(400, 'res_err_sign_up_otp_not_found', 'OTP not found.');
        }
        return _signedIn();
      case ('POST', ApiEndpoints.refreshToken):
        return _error(401, 'res_err_unauthorized', 'Refresh token expired');
      case ('DELETE', ApiEndpoints.logout):
        return ResponseBody.fromString(
          '',
          204,
          headers: {
            'set-cookie': ['ACCESS_TOKEN=; Path=/; Max-Age=0'],
          },
        );
    }
    return _error(404, 'res_err_route_not_found', 'Not found');
  }
}

/// Handles to the pumped app for assertions.
class TestApp {
  const TestApp({
    required this.server,
    required this.adapter,
    required this.jar,
    required this.container,
  });

  final FakeAuthServer server;
  final FakeAdapter adapter;
  final CookieJar jar;
  final ProviderContainer container;

  /// Requests sent to [path], in order.
  List<RequestOptions> requestsTo(String path) =>
      adapter.requests.where((r) => r.path == path).toList();

  Future<bool> hasSessionCookie() async =>
      (await jar.loadForRequest(Uri.parse(testBaseUrl)))
          .any((c) => c.name == 'ACCESS_TOKEN' && c.value.isNotEmpty);
}

/// Pumps the real [LetsLiveApp] (router, session, screens) against
/// [FakeAuthServer]. [signedIn] starts with a session cookie in the jar.
/// App Check tokens are numbered so tests can see a fresh one per request.
/// Pass `settle: false` when the first screen keeps animating (the splash),
/// [jar] to swap the in-memory cookie jar (e.g. for a failing one), and
/// [overrides] for extra provider overrides.
Future<TestApp> pumpTestApp(
  WidgetTester tester, {
  FakeAuthServer? server,
  bool signedIn = false,
  GoogleSignInService? google,
  bool settle = true,
  Map<String, Object> preferences = const {},
  CookieJar? jar,
  List<Override> overrides = const [],
}) async {
  SharedPreferences.setMockInitialValues(preferences);
  final prefs = await SharedPreferences.getInstance();
  final fakeServer = server ?? FakeAuthServer();
  final adapter = FakeAdapter(fakeServer.handle);
  final cookies = jar ?? CookieJar();
  if (signedIn) {
    await cookies.saveFromResponse(Uri.parse(testBaseUrl), [
      Cookie('ACCESS_TOKEN', 'access'),
      Cookie('REFRESH_TOKEN', 'refresh'),
    ]);
  }
  final client = ApiClient(
    baseUrl: testBaseUrl,
    cookieJar: cookies,
    adapter: adapter,
  );
  addTearDown(client.dispose);
  var appCheckTokens = 0;

  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        sharedPreferencesProvider.overrideWithValue(prefs),
        appConfigProvider.overrideWithValue(testConfig),
        cookieJarProvider.overrideWithValue(cookies),
        apiClientProvider.overrideWithValue(client),
        botProofProvider.overrideWithValue(
          BotProof.appCheck(() async => 'app-check-${++appCheckTokens}'),
        ),
        googleSignInServiceProvider.overrideWithValue(google),
        ...overrides,
      ],
      child: const LetsLiveApp(),
    ),
  );
  if (settle) await tester.pumpAndSettle();

  return TestApp(
    server: fakeServer,
    adapter: adapter,
    jar: cookies,
    container: ProviderScope.containerOf(
      tester.element(find.byType(LetsLiveApp)),
    ),
  );
}

/// Finds the text field with [key] and types [text].
Future<void> enter(WidgetTester tester, String key, String text) =>
    tester.enterText(find.byKey(ValueKey(key)), text);
