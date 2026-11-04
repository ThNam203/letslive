import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/network/secure_cookie_storage.dart';

import 'fake_adapter.dart';

const _base = 'http://localhost:8000';

String _parseName(Object? data) =>
    (data! as Map<String, Object?>)['name']! as String;

ApiClient _client(FakeAdapter adapter, {CookieJar? jar}) =>
    ApiClient(baseUrl: _base, cookieJar: jar ?? CookieJar(), adapter: adapter);

void main() {
  test('parses success envelope with meta', () async {
    final adapter = FakeAdapter(
      (_) => okEnvelope(
        {'name': 'bob'},
        meta: {'page': 2, 'page_size': 10, 'total': 55},
      ),
    );
    final res = await _client(adapter)
        .get<String>('/user/me', parse: _parseName);

    expect(res.success, isTrue);
    expect(res.statusCode, 200);
    expect(res.data, 'bob');
    expect(res.requestId, 'req-1');
    expect(res.meta?.page, 2);
    expect(res.meta?.pageSize, 10);
    expect(res.meta?.total, 55);
    expect(adapter.requests.single.uri.toString(), '$_base/user/me');
  });

  test(
    'maps error envelope to ApiResponse(success:false) not exception',
    () async {
      final adapter = FakeAdapter(
        (_) => jsonBody({
          'requestId': 'req-9',
          'success': false,
          'code': 1001,
          'key': 'invalid_input',
          'message': 'bad',
          'errorDetails': [
            {'field': 'email', 'reason': 'required'},
            'garbage',
          ],
        }, status: 400),
      );
      final res = await _client(adapter)
          .post<String>('/auth/login', body: {'a': 1}, parse: _parseName);

      expect(res.success, isFalse);
      expect(res.statusCode, 400);
      expect(res.code, 1001);
      expect(res.key, 'invalid_input');
      expect(res.errorDetails, [
        {'field': 'email', 'reason': 'required'},
      ]);
      expect(res.data, isNull);
      expect(res.failure, isNull);
      expect(res.isRetryable, isFalse);
    },
  );

  test('malformed body yields a typed failure instead of crashing', () async {
    final adapter = FakeAdapter((_) => ResponseBody.fromString('<html>', 502));
    final res = await _client(adapter)
        .get<String>('/user/me', parse: _parseName);

    expect(res.success, isFalse);
    expect(res.failure, ApiFailureKind.malformedResponse);
    expect(res.statusCode, 502);
    expect(res.isRetryable, isTrue); // 502 is a gateway error
  });

  test('parse failure on success data yields malformedResponse', () async {
    final adapter = FakeAdapter((_) => okEnvelope({'nope': 1}));
    final res = await _client(adapter)
        .get<String>('/user/me', parse: _parseName);

    expect(res.success, isFalse);
    expect(res.failure, ApiFailureKind.malformedResponse);
  });

  test(
    '401 triggers one refresh for N concurrent requests then retries each once',
    () async {
      var refreshed = false;
      final adapter = FakeAdapter((o) async {
        if (o.path == ApiEndpoints.refreshToken) {
          await Future<void>.delayed(const Duration(milliseconds: 50));
          refreshed = true;
          return okEnvelope(null);
        }
        if (!refreshed) {
          return jsonBody({'success': false, 'code': 401}, status: 401);
        }
        return okEnvelope({'name': o.path});
      });
      final client = _client(adapter);

      final results = await Future.wait([
        for (var i = 0; i < 5; i++)
          client.get<String>('/user/me', parse: _parseName),
      ]);

      expect(adapter.count(ApiEndpoints.refreshToken), 1);
      expect(adapter.count('/user/me'), 10); // 5 originals + 5 retries
      expect(results.every((r) => r.success && r.data == '/user/me'), isTrue);
    },
  );

  test(
    'refreshSession shares the in-flight refresh with concurrent callers',
    () async {
      final adapter = FakeAdapter((o) async {
        await Future<void>.delayed(const Duration(milliseconds: 20));
        return okEnvelope(null);
      });
      final client = _client(adapter);

      final results = await Future.wait([
        client.refreshSession(),
        client.refreshSession(),
        client.refreshSession(),
      ]);

      expect(results, [true, true, true]);
      expect(adapter.count(ApiEndpoints.refreshToken), 1);
    },
  );

  test('refresh failure emits sessionExpired and does not loop', () async {
    final adapter = FakeAdapter((o) {
      if (o.path == ApiEndpoints.refreshToken) {
        return jsonBody({'success': false, 'code': 401}, status: 401);
      }
      return jsonBody({
        'success': false,
        'code': 401,
        'key': 'unauth',
      }, status: 401);
    });
    final client = _client(adapter);
    var expired = 0;
    final sub = client.sessionExpired.listen((_) => expired++);

    final results = await Future.wait([
      client.get<String>('/user/me', parse: _parseName),
      client.get<String>('/user/me', parse: _parseName),
    ]);
    await Future<void>.delayed(Duration.zero);

    expect(adapter.count(ApiEndpoints.refreshToken), 1);
    expect(adapter.count('/user/me'), 2); // no retry
    expect(expired, 1);
    expect(results.every((r) => !r.success && r.statusCode == 401), isTrue);
    await sub.cancel();
  });

  test(
    'a replayed request that is 401 again does not refresh a second time',
    () async {
      final adapter = FakeAdapter((o) {
        if (o.path == ApiEndpoints.refreshToken) return okEnvelope(null);
        return jsonBody({'success': false, 'code': 401}, status: 401);
      });
      final res = await _client(adapter)
          .get<String>('/user/me', parse: _parseName);

      expect(res.statusCode, 401);
      expect(adapter.count(ApiEndpoints.refreshToken), 1);
      expect(adapter.count('/user/me'), 2);
    },
  );

  test('401 on auth paths does not trigger a refresh', () async {
    final adapter = FakeAdapter(
      (_) => jsonBody({'success': false, 'code': 401}, status: 401),
    );
    final res = await _client(adapter)
        .post<String>('/auth/login', body: {}, parse: _parseName);

    expect(res.statusCode, 401);
    expect(adapter.count(ApiEndpoints.refreshToken), 0);
  });

  test('network failure during refresh does not emit sessionExpired', () async {
    final adapter = FakeAdapter((o) {
      if (o.path == ApiEndpoints.refreshToken) {
        throw DioException.connectionError(requestOptions: o, reason: 'down');
      }
      return jsonBody({'success': false}, status: 401);
    });
    final client = _client(adapter);
    var expired = 0;
    final sub = client.sessionExpired.listen((_) => expired++);

    final ok = await client.refreshSession();
    await Future<void>.delayed(Duration.zero);

    expect(ok, isFalse);
    expect(expired, 0);
    await sub.cancel();
  });

  test('network timeout yields a retryable error', () async {
    final adapter = FakeAdapter(
      (o) => throw DioException.receiveTimeout(
        timeout: const Duration(seconds: 15),
        requestOptions: o,
      ),
    );
    final res = await _client(adapter)
        .get<String>('/user/me', parse: _parseName);

    expect(res.success, isFalse);
    expect(res.failure, ApiFailureKind.timeout);
    expect(res.isRetryable, isTrue);
    expect(res.statusCode, 0);
  });

  test('connection error yields a retryable network failure', () async {
    final adapter = FakeAdapter(
      (o) => throw DioException.connectionError(
        requestOptions: o,
        reason: 'offline',
      ),
    );
    final res = await _client(adapter)
        .get<String>('/user/me', parse: _parseName);

    expect(res.failure, ApiFailureKind.network);
    expect(res.isRetryable, isTrue);
  });

  test('cancelled request yields a non-retryable failure', () async {
    final adapter = FakeAdapter(
      (o) => throw DioException.requestCancelled(
        requestOptions: o,
        reason: 'user',
      ),
    );
    final res = await _client(adapter)
        .get<String>('/user/me', parse: _parseName);

    expect(res.failure, ApiFailureKind.cancelled);
    expect(res.isRetryable, isFalse);
  });

  test(
    'upload sends multipart, reports progress and replays after refresh',
    () async {
      var refreshed = false;
      final adapter = FakeAdapter((o) async {
        if (o.path == ApiEndpoints.refreshToken) {
          refreshed = true;
          return okEnvelope(null);
        }
        if (!refreshed) {
          return jsonBody({'success': false}, status: 401);
        }
        expect(o.data, isA<FormData>());
        return okEnvelope({'name': 'uploaded'});
      });
      final form = FormData.fromMap({
        'file': MultipartFile.fromBytes([1, 2, 3], filename: 'a.bin'),
      });
      final res = await _client(adapter)
          .upload<String>('/upload/avatar', formData: form, parse: _parseName);

      expect(res.success, isTrue);
      expect(res.data, 'uploaded');
      expect(adapter.count('/upload/avatar'), 2);
      // uploads are not subject to the default timeout
      expect(adapter.requests.first.sendTimeout, Duration.zero);
      expect(adapter.requests.first.receiveTimeout, Duration.zero);
    },
  );

  test('default requests use a 15s timeout', () async {
    final adapter = FakeAdapter((_) => okEnvelope({'name': 'x'}));
    await _client(adapter).get<String>('/user/me', parse: _parseName);

    expect(adapter.requests.single.receiveTimeout, const Duration(seconds: 15));
    expect(adapter.requests.single.connectTimeout, const Duration(seconds: 15));
  });

  group('cookies', () {
    const cookieHeader = 'set-cookie';

    test('cookie jar persists across client instances', () async {
      FlutterSecureStorage.setMockInitialValues({});
      final loginAdapter = FakeAdapter(
        (_) => jsonBody(
          {'success': true},
          headers: {
            cookieHeader: [
              'ACCESS_TOKEN=abc; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=None',
              'REFRESH_TOKEN=def; Path=/; Max-Age=86400; HttpOnly; Secure; SameSite=None',
            ],
          },
        ),
      );
      final first = _client(loginAdapter, jar: createSecureCookieJar());
      await first.post<String>('/auth/login', body: {}, parse: _parseName);

      // New process: fresh jar + client over the same secure storage.
      final meAdapter = FakeAdapter((_) => okEnvelope({'name': 'bob'}));
      final second = _client(meAdapter, jar: createSecureCookieJar());
      await second.get<String>('/user/me', parse: _parseName);

      final sent = meAdapter.requests.single.headers['cookie'] as String?;
      expect(sent, contains('ACCESS_TOKEN=abc'));
      expect(sent, contains('REFRESH_TOKEN=def'));
    });

    test('tokens are written to secure storage, not files', () async {
      FlutterSecureStorage.setMockInitialValues({});
      final adapter = FakeAdapter(
        (_) => jsonBody(
          {'success': true},
          headers: {
            cookieHeader: ['ACCESS_TOKEN=abc; Path=/; Max-Age=3600'],
          },
        ),
      );
      await _client(
        adapter,
        jar: createSecureCookieJar(),
      ).post<String>('/auth/login', body: {}, parse: _parseName);

      final stored = await const FlutterSecureStorage().readAll();
      expect(stored.keys, isNotEmpty);
      expect(
        stored.keys.every((k) => k.startsWith('letslive.cookies.')),
        isTrue,
      );
      expect(stored.values.any((v) => v.contains('abc')), isTrue);
    });
  });
}
