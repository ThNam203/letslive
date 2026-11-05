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

  test('default headers are sent on every request', () async {
    final adapter = FakeAdapter((_) => okEnvelope(null));
    final client = ApiClient(
      baseUrl: _base,
      cookieJar: CookieJar(),
      adapter: adapter,
      defaultHeaders: const {'User-Agent': 'contract-tests'},
    );

    await client.get<Object?>('/a', parse: _noData);
    await client.get<Object?>('/b', parse: _noData);

    for (final request in adapter.requests) {
      expect(request.headers['User-Agent'], 'contract-tests');
    }
  });

  test('a per-call header wins over the default of the same name', () async {
    final adapter = FakeAdapter((_) => okEnvelope(null));
    final client = ApiClient(
      baseUrl: _base,
      cookieJar: CookieJar(),
      adapter: adapter,
      defaultHeaders: const {'X-Token': 'default'},
    );

    await client.get<Object?>(
      '/a',
      parse: _noData,
      headers: {'X-Token': 'from-call'},
    );

    expect(adapter.requests.single.headers['X-Token'], 'from-call');
  });

  test('default headers also reach the refresh call and the replay', () async {
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
      defaultHeaders: const {'X-Token': 't'},
    );

    final res = await client.get<Object?>('/user/me', parse: _noData);

    expect(res.success, isTrue);
    expect(adapter.requests, hasLength(3));
    for (final request in adapter.requests) {
      expect(request.headers['X-Token'], 't', reason: request.path);
    }
  });
}
