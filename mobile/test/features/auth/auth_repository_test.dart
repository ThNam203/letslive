import 'dart:convert';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';

import '../../core/network/fake_adapter.dart';

const _base = 'http://localhost:8000';

Map<String, Object?> bodyOf(RequestOptions request) =>
    jsonDecode(jsonEncode(request.data)) as Map<String, Object?>;

void main() {
  late FakeAdapter adapter;
  late ApiClient client;

  setUp(() {
    adapter = FakeAdapter((_) => okEnvelope(null));
    client = ApiClient(
      baseUrl: _base,
      cookieJar: CookieJar(),
      adapter: adapter,
    );
  });

  tearDown(() => client.dispose());

  group('App Check bot proof', () {
    late int tokens;
    late AuthRepository repo;

    setUp(() {
      tokens = 0;
      repo = AuthRepository(
        client: client,
        botProof: BotProof.appCheck(() async => 'app-check-${++tokens}'),
      );
    });

    test('login sends a fresh App Check token on every attempt', () async {
      await repo.login('a@b.co', 'secret-1');
      await repo.login('a@b.co', 'secret-1');

      final first = adapter.requests[0];
      expect(first.path, ApiEndpoints.login);
      expect(first.method, 'POST');
      expect(first.headers['X-Firebase-AppCheck'], 'app-check-1');
      expect(adapter.requests[1].headers['X-Firebase-AppCheck'], 'app-check-2');
      expect(bodyOf(first), {'email': 'a@b.co', 'password': 'secret-1'});
    });

    test('email verification carries the App Check token', () async {
      await repo.requestEmailVerification('a@b.co');

      final request = adapter.requests.single;
      expect(request.path, ApiEndpoints.verifyEmail);
      expect(request.headers['X-Firebase-AppCheck'], 'app-check-1');
      expect(bodyOf(request), {'email': 'a@b.co'});
    });

    test('a failing token fetch is a retryable failure, not a crash', () async {
      final failing = AuthRepository(
        client: client,
        botProof: BotProof.appCheck(
          () async => throw Exception('play integrity unavailable'),
        ),
      );

      final res = await failing.login('a@b.co', 'secret-1');

      expect(res.success, isFalse);
      expect(res.isRetryable, isTrue);
      expect(adapter.requests, isEmpty);
    });
  });

  group('Turnstile dev bot proof', () {
    test('login puts the dev token in the body and sends no header', () async {
      final repo = AuthRepository(
        client: client,
        botProof: const BotProof.turnstile('dev-token'),
      );

      await repo.login('a@b.co', 'secret-1');

      final request = adapter.requests.single;
      expect(request.headers.containsKey('X-Firebase-AppCheck'), isFalse);
      expect(bodyOf(request), {
        'email': 'a@b.co',
        'password': 'secret-1',
        'turnstileToken': 'dev-token',
      });
    });
  });

  group('endpoints without bot proof', () {
    late AuthRepository repo;

    setUp(() {
      repo = AuthRepository(
        client: client,
        botProof: BotProof.appCheck(() async => 'unused'),
      );
    });

    test('google sign in posts id token', () async {
      final res = await repo.loginWithGoogle('google-id-token');

      final request = adapter.requests.single;
      expect(res.success, isTrue);
      expect(request.method, 'POST');
      expect(request.path, '/auth/google/mobile');
      expect(bodyOf(request), {'idToken': 'google-id-token'});
      expect(request.headers.containsKey('X-Firebase-AppCheck'), isFalse);
    });

    test('signup posts the form and the otp code', () async {
      await repo.signup(
        email: 'a@b.co',
        username: 'abcdef',
        password: 'Aa!xxxxx',
        otpCode: '123456',
      );

      final request = adapter.requests.single;
      expect(request.path, ApiEndpoints.signup);
      expect(bodyOf(request), {
        'email': 'a@b.co',
        'username': 'abcdef',
        'password': 'Aa!xxxxx',
        'otpCode': '123456',
      });
    });

    test('logout is DELETE /auth/logout', () async {
      await repo.logout();

      final request = adapter.requests.single;
      expect(request.method, 'DELETE');
      expect(request.path, ApiEndpoints.logout);
    });

    test('server errors come back as values', () async {
      adapter.handler = (_) => jsonBody({
        'requestId': 'r',
        'success': false,
        'code': 20007,
        'key': 'res_err_email_or_password_incorrect',
        'message': 'Username or password incorrect.',
      }, status: 401);

      final ApiResponse<void> res = await repo.login('a@b.co', 'wrong-pass');

      expect(res.success, isFalse);
      expect(res.key, 'res_err_email_or_password_incorrect');
    });
  });
}
