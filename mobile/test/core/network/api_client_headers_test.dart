import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

import 'fake_adapter.dart';

const _base = 'http://localhost:8000';

Object? _noData(Object? data) => data;

CookieJar _loggedInJar() {
  final jar = CookieJar();
  jar.saveFromResponse(Uri.parse(_base), [
    Cookie('REFRESH_TOKEN', 'r'),
    Cookie('ACCESS_TOKEN', 'a'),
  ]);
  return jar;
}

void main() {
  test('per-call headers are sent on that request only', () async {
    final adapter = FakeAdapter((_) => okEnvelope(null));
    final client = ApiClient(
      baseUrl: _base,
      cookieJar: CookieJar(),
      adapter: adapter,
    );

    await client.post<Object?>(
      ApiEndpoints.login,
      parse: _noData,
      headers: {'X-Extra': 'one'},
    );
    await client.get<Object?>('/user/me', parse: _noData);

    expect(adapter.requests[0].headers['X-Extra'], 'one');
    expect(adapter.requests[1].headers.containsKey('X-Extra'), isFalse);
  });

  test('headerProvider headers are sent on every request', () async {
    var calls = 0;
    final adapter = FakeAdapter((_) => okEnvelope(null));
    final client = ApiClient(
      baseUrl: _base,
      cookieJar: CookieJar(),
      adapter: adapter,
      headerProvider: () async => {'X-Token': 'token-${++calls}'},
    );

    await client.get<Object?>('/a', parse: _noData);
    await client.get<Object?>('/b', parse: _noData);

    expect(adapter.requests[0].headers['X-Token'], 'token-1');
    expect(adapter.requests[1].headers['X-Token'], 'token-2');
  });

  test(
    'per-call header wins over the provider header of the same name',
    () async {
      final adapter = FakeAdapter((_) => okEnvelope(null));
      final client = ApiClient(
        baseUrl: _base,
        cookieJar: CookieJar(),
        adapter: adapter,
        headerProvider: () => {'X-Token': 'from-provider'},
      );

      await client.get<Object?>(
        '/a',
        parse: _noData,
        headers: {'x-token': 'from-call'},
      );

      expect(adapter.requests.single.headers['X-Token'], 'from-call');
    },
  );

  test('provider headers also reach the refresh call and the replay', () async {
    var refreshed = false;
    final adapter = FakeAdapter((o) {
      if (o.path == ApiEndpoints.refreshToken) {
        refreshed = true;
        return okEnvelope(null);
      }
      return refreshed ? okEnvelope(null) : jsonBody({}, status: 401);
    });
    final client = ApiClient(
      baseUrl: _base,
      cookieJar: _loggedInJar(),
      adapter: adapter,
      headerProvider: () => {'X-Token': 't'},
    );

    final res = await client.get<Object?>('/user/me', parse: _noData);

    expect(res.success, isTrue);
    expect(adapter.requests, hasLength(3));
    for (final request in adapter.requests) {
      expect(request.headers['X-Token'], 't', reason: request.path);
    }
  });

  for (final name in ['Cookie', 'authorization']) {
    // A programming error, not a (retryable) network failure.
    test('headerProvider setting $name throws a StateError', () async {
      final adapter = FakeAdapter((_) => okEnvelope(null));
      final client = ApiClient(
        baseUrl: _base,
        cookieJar: _loggedInJar(),
        adapter: adapter,
        headerProvider: () => {name: 'forged'},
      );

      await expectLater(
        client.get<Object?>('/user/me', parse: _noData),
        throwsA(
          isA<StateError>().having(
            (e) => e.message,
            'message',
            contains(name.toLowerCase()),
          ),
        ),
      );
      expect(adapter.requests, isEmpty, reason: 'request must not be sent');
    });
  }
}
