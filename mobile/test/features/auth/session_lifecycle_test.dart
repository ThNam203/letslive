import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';
import 'package:letslive_mobile/features/auth/domain/session_lifecycle.dart';

import '../../core/network/fake_adapter.dart';

const _base = 'http://localhost:8000';
final _baseUri = Uri.parse(_base);

void main() {
  late CookieJar jar;
  late FakeAdapter adapter;
  late ApiClient client;
  late List<String> events;
  late SessionLifecycle lifecycle;

  Future<bool> jarHasCookies() async =>
      (await jar.loadForRequest(_baseUri)).isNotEmpty;

  setUp(() async {
    jar = CookieJar();
    await jar.saveFromResponse(_baseUri, [
      Cookie('ACCESS_TOKEN', 'a'),
      Cookie('REFRESH_TOKEN', 'r'),
    ]);
    events = [];
    adapter = FakeAdapter((options) {
      events.add('${options.method} ${options.path}');
      return noContent();
    });
    client = ApiClient(baseUrl: _base, cookieJar: jar, adapter: adapter);
    lifecycle = SessionLifecycle(
      auth: AuthRepository(
        client: client,
        botProof: const BotProof.turnstile('t'),
      ),
      cookieJar: jar,
      restartRealtime: () => events.add('realtime restart'),
    );
  });

  tearDown(() => client.dispose());

  test('logout clears cookie jar and unregisters device token', () async {
    // Stands in for PushService.unregister (Task 17): it must run while the
    // session cookies still exist, so it can authenticate.
    lifecycle.addLogoutHook(() async {
      events.add('unregister device (cookies: ${await jarHasCookies()})');
    });

    await lifecycle.endSession(notifyServer: true);

    expect(events, [
      'unregister device (cookies: true)',
      'DELETE ${ApiEndpoints.logout}',
      'realtime restart',
    ]);
    expect(await jarHasCookies(), isFalse);
  });

  test('a failing or hanging hook never blocks logout', () async {
    lifecycle
      ..addLogoutHook(() async => throw StateError('push service down'))
      ..addLogoutHook(() => Completer<void>().future)
      ..addLogoutHook(() async => events.add('last hook'));

    await lifecycle.endSession(
      notifyServer: true,
      hookTimeout: const Duration(milliseconds: 10),
    );

    expect(events, [
      'last hook',
      'DELETE ${ApiEndpoints.logout}',
      'realtime restart',
    ]);
    expect(await jarHasCookies(), isFalse);
  });

  test('a failed logout call still clears the local session', () async {
    adapter.handler = (_) => throw StateError('offline');

    await lifecycle.endSession(notifyServer: true);

    expect(await jarHasCookies(), isFalse);
    expect(events, ['realtime restart']);
  });

  test('a failed cookie wipe still restarts realtime, then reports', () async {
    final failing = SessionLifecycle(
      auth: AuthRepository(
        client: client,
        botProof: const BotProof.turnstile('t'),
      ),
      cookieJar: _FailingDeleteJar(),
      restartRealtime: () => events.add('realtime restart'),
    );

    await expectLater(
      failing.endSession(notifyServer: false),
      throwsA(isA<StateError>()),
    );
    expect(events, ['realtime restart']);
  });

  test('session expiry skips the server call', () async {
    await lifecycle.endSession(notifyServer: false);

    expect(events, ['realtime restart']);
    expect(await jarHasCookies(), isFalse);
  });

  test('concurrent endSession calls share one run (no expiry loop)', () async {
    var hookRuns = 0;
    lifecycle.addLogoutHook(() async {
      hookRuns++;
      // A hook whose request fails auth can make the client report another
      // expiry while logout is still running.
      unawaited(lifecycle.endSession(notifyServer: false));
    });

    await Future.wait([
      lifecycle.endSession(notifyServer: true),
      lifecycle.endSession(notifyServer: false),
    ]);

    expect(hookRuns, 1);
    expect(events.where((e) => e == 'realtime restart'), hasLength(1));
  });

  test('signedIn restarts the realtime socket', () {
    lifecycle.signedIn();

    expect(events, ['realtime restart']);
  });
}

/// A 204 with no body, like the real `DELETE /auth/logout`.
ResponseBody noContent() => ResponseBody.fromString('', 204);

/// A cookie jar whose wipe fails, like a secure-storage write error.
class _FailingDeleteJar extends DefaultCookieJar {
  @override
  Future<void> deleteAll() async => throw StateError('storage unavailable');
}
