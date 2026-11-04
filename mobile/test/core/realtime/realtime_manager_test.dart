import 'package:cookie_jar/cookie_jar.dart';
import 'package:fake_async/fake_async.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';

import 'fake_channel.dart';

const _wsUrl = 'ws://localhost:8000/realtime';

class _Harness {
  _Harness({
    Map<String, String> cookies = const {
      'ACCESS_TOKEN': 'a',
      'REFRESH_TOKEN': 'r',
    },
    this.refreshResult = true,
  }) : jar = CookieJar() {
    jar.saveFromResponse(Uri.parse('http://localhost:8000'), [
      for (final e in cookies.entries) Cookie(e.key, e.value),
    ]);
    manager = RealtimeManager(
      url: _wsUrl,
      cookieJar: jar,
      refreshSession: () async {
        refreshCalls++;
        return refreshResult;
      },
      connect: (uri, headers) {
        final channel = FakeChannel(uri, headers);
        channels.add(channel);
        return channel;
      },
    );
  }

  final CookieJar jar;
  late final RealtimeManager manager;
  final List<FakeChannel> channels = [];
  final List<RealtimeStatus> statuses = [];
  bool refreshResult;
  int refreshCalls = 0;

  FakeChannel get last => channels.last;
}

/// Starts the manager and completes the first handshake.
void _startOpen(_Harness h, FakeAsync async) {
  h.manager.status.listen(h.statuses.add);
  h.manager.start();
  async.flushMicrotasks();
  h.last.open();
  async.flushMicrotasks();
}

void main() {
  group('RealtimeManager', () {
    test('connects with the jar cookies as a Cookie header', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        expect(h.last.uri.toString(), _wsUrl);
        final cookie = h.last.headers['Cookie']!;
        expect(cookie, contains('ACCESS_TOKEN=a'));
        expect(cookie, contains('REFRESH_TOKEN=r'));
        expect(h.statuses.last, RealtimeStatus.connected);
      });
    });

    test('second subscriber to same topic sends no extra frame', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        h.manager.subscribe('room:1');
        h.manager.subscribe('room:1');
        expect(h.last.opsFor('subscribe'), ['room:1']);
      });
    });

    test('last unsubscribe sends unsubscribe', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        final a = h.manager.subscribe('room:1');
        final b = h.manager.subscribe('room:1');
        a();
        a(); // idempotent
        expect(h.last.opsFor('unsubscribe'), isEmpty);
        b();
        expect(h.last.opsFor('unsubscribe'), ['room:1']);
      });
    });

    test('subscribing before the socket opens sends once on open', () {
      fakeAsync((async) {
        final h = _Harness();
        h.manager.start();
        h.manager.subscribe('room:1');
        async.flushMicrotasks();
        h.last.open();
        async.flushMicrotasks();
        expect(h.last.opsFor('subscribe'), ['room:1']);
      });
    });

    test('reconnect resubscribes every topic and fires onReconnect', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        var reconnects = 0;
        h.manager.onReconnect(() => reconnects++);
        h.manager.subscribe('room:1');
        h.manager.subscribe('room:2');

        h.last.serverCloses(1006);
        async.flushMicrotasks();
        expect(h.statuses.last, RealtimeStatus.reconnecting);
        async.elapse(const Duration(seconds: 1));
        h.last.open();
        async.flushMicrotasks();

        expect(h.channels, hasLength(2));
        expect(
          h.last.opsFor('subscribe'),
          unorderedEquals(['room:1', 'room:2']),
        );
        expect(reconnects, 1);
        expect(h.statuses.last, RealtimeStatus.connected);
      });
    });

    test('first connect does not fire onReconnect', () {
      fakeAsync((async) {
        final h = _Harness();
        var reconnects = 0;
        h.manager.onReconnect(() => reconnects++);
        _startOpen(h, async);
        expect(reconnects, 0);
      });
    });

    test('backoff doubles up to 30s and resets after a successful open', () {
      fakeAsync((async) {
        final h = _Harness();
        h.manager.start();
        async.flushMicrotasks();
        final expected = [1, 2, 4, 8, 16, 30, 30];
        for (final seconds in expected) {
          final before = h.channels.length;
          h.last.failHandshake(null);
          async.flushMicrotasks();
          async.elapse(
            Duration(seconds: seconds) - const Duration(milliseconds: 1),
          );
          expect(
            h.channels.length,
            before,
            reason: 'too early for ${seconds}s',
          );
          async.elapse(const Duration(milliseconds: 1));
          async.flushMicrotasks();
          expect(h.channels.length, before + 1, reason: '${seconds}s retry');
        }
        // success resets the delay to 1s
        h.last.open();
        async.flushMicrotasks();
        h.last.serverCloses(1006);
        async.flushMicrotasks();
        final before = h.channels.length;
        async.elapse(const Duration(seconds: 1));
        async.flushMicrotasks();
        expect(h.channels.length, before + 1);
      });
    });

    test('pause closes socket on background and resume reconnects', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        var reconnects = 0;
        h.manager.onReconnect(() => reconnects++);
        h.manager.subscribe('room:1');
        final first = h.last;

        h.manager.pause();
        async.flushMicrotasks();
        expect(first.closedByClient, isTrue);
        expect(h.statuses.last, RealtimeStatus.paused);
        async.elapse(const Duration(minutes: 5));
        expect(h.channels, hasLength(1), reason: 'no reconnect while paused');

        h.manager.resume();
        async.flushMicrotasks();
        expect(h.channels, hasLength(2));
        h.last.open();
        async.flushMicrotasks();
        expect(h.last.opsFor('subscribe'), ['room:1']);
        expect(reconnects, 1);
      });
    });

    test('unauthorized close triggers refresh once then reconnect', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        h.manager.subscribe('room:1');

        h.last.serverCloses(1008);
        async.flushMicrotasks();
        expect(h.refreshCalls, 1);
        // refresh succeeded: reconnects immediately, no backoff wait
        expect(h.channels, hasLength(2));
        h.last.open();
        async.flushMicrotasks();
        expect(h.last.opsFor('subscribe'), ['room:1']);
        expect(h.statuses.last, RealtimeStatus.connected);
      });
    });

    test(
      'handshake 401 refreshes once; a second 401 stops as unauthorized',
      () {
        fakeAsync((async) {
          final h = _Harness();
          h.manager.status.listen(h.statuses.add);
          h.manager.start();
          async.flushMicrotasks();
          h.last.failHandshake(401);
          async.flushMicrotasks();
          expect(h.refreshCalls, 1);
          expect(h.channels, hasLength(2));

          h.last.failHandshake(401);
          async.flushMicrotasks();
          async.elapse(const Duration(minutes: 5));
          expect(h.refreshCalls, 1, reason: 'no refresh loop');
          expect(h.channels, hasLength(2), reason: 'no reconnect loop');
          expect(h.statuses.last, RealtimeStatus.unauthorized);
        });
      },
    );

    test('failed refresh stops reconnecting with unauthorized status', () {
      fakeAsync((async) {
        final h = _Harness(refreshResult: false);
        _startOpen(h, async);
        h.last.serverCloses(1008);
        async.flushMicrotasks();
        async.elapse(const Duration(minutes: 5));
        expect(h.refreshCalls, 1);
        expect(h.channels, hasLength(1));
        expect(h.statuses.last, RealtimeStatus.unauthorized);

        // start() after a fresh login resets the state
        h.refreshResult = true;
        h.manager.start();
        async.flushMicrotasks();
        expect(h.channels, hasLength(2));
      });
    });

    test(
      'refreshes before connecting when only the refresh cookie is left',
      () {
        fakeAsync((async) {
          final h = _Harness(cookies: {'REFRESH_TOKEN': 'r'});
          h.manager.start();
          async.flushMicrotasks();
          expect(h.refreshCalls, 1);
          expect(h.channels, hasLength(1));
        });
      },
    );

    test('guest connects anonymously without refreshing', () {
      fakeAsync((async) {
        final h = _Harness(cookies: const {});
        h.manager.start();
        async.flushMicrotasks();
        expect(h.refreshCalls, 0);
        expect(h.channels, hasLength(1));
        expect(h.last.headers.containsKey('Cookie'), isFalse);
      });
    });

    test('dispatches events by type and ignores unknown or malformed', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        final got = <Object?>[];
        final off = h.manager.onEvent('chat.message', got.add);
        h.last.serverSends({
          'op': 'event',
          'topic': 'room:1',
          'type': 'chat.message',
          'data': {'x': 1},
        });
        h.last.serverSends({
          'op': 'event',
          'topic': 'room:1',
          'type': 'other',
          'data': 1,
        });
        h.last.serverSendsRaw('garbage');
        h.last.serverSends({'op': 'error', 'code': 'forbidden_topic'});
        async.flushMicrotasks();
        expect(got, [
          {'x': 1},
        ]);

        off();
        h.last.serverSends({
          'op': 'event',
          'topic': 'room:1',
          'type': 'chat.message',
          'data': 2,
        });
        async.flushMicrotasks();
        expect(got, hasLength(1));
      });
    });

    test('a throwing handler does not stop other handlers', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        var called = 0;
        h.manager.onEvent('chat.message', (_) => throw Exception('boom'));
        h.manager.onEvent('chat.message', (_) => called++);
        h.last.serverSends({
          'op': 'event',
          'topic': 'room:1',
          'type': 'chat.message',
        });
        async.flushMicrotasks();
        expect(called, 1);
      });
    });

    test('sends an app ping periodically and reconnects when no reply', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        async.elapse(RealtimeManager.defaultPingInterval);
        expect(h.last.opsFor('ping'), hasLength(1));
        h.last.serverSends({'op': 'pong'});
        async.flushMicrotasks();
        async.elapse(RealtimeManager.defaultPingInterval);
        expect(h.last.opsFor('ping'), hasLength(2));

        // no pong: the socket is considered dead
        final dead = h.last;
        async.elapse(RealtimeManager.defaultPongTimeout);
        async.flushMicrotasks();
        expect(dead.closedByClient, isTrue);
        async.elapse(const Duration(seconds: 1));
        async.flushMicrotasks();
        expect(h.channels, hasLength(2));
      });
    });

    test('dispose closes the socket and stops reconnecting', () {
      fakeAsync((async) {
        final h = _Harness();
        _startOpen(h, async);
        h.manager.dispose();
        async.flushMicrotasks();
        expect(h.last.closedByClient, isTrue);
        async.elapse(const Duration(minutes: 5));
        expect(h.channels, hasLength(1));
      });
    });
  });
}
