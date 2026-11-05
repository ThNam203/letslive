import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/features/watch/chat/chat_providers.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';
import 'package:letslive_mobile/features/watch/chat/domain/chat_line.dart';

import '../../../core/realtime/fake_channel.dart';
import 'chat_fakes.dart';

const _room = 'room-1';

/// A real [RealtimeManager] over [FakeChannel]s, plus the provider container
/// the chat controller lives in.
class _Rig {
  _Rig(this.repo) {
    manager = RealtimeManager(
      url: 'ws://localhost:8000/realtime',
      cookieJar: CookieJar(),
      refreshSession: () async => true,
      connect: (uri, headers) {
        final channel = FakeChannel(uri, headers);
        channels.add(channel);
        return channel;
      },
    );
    container = ProviderContainer(
      overrides: [
        realtimeManagerProvider.overrideWithValue(manager),
        chatRepositoryProvider.overrideWithValue(repo),
      ],
    );
  }

  final FakeChatRepository repo;
  late final RealtimeManager manager;
  late final ProviderContainer container;
  final List<FakeChannel> channels = [];
  ProviderSubscription<AsyncValue<List<ChatLine>>>? _listener;

  FakeChannel get channel => channels.last;

  /// Connects the socket (handshake done) and opens the chat controller.
  Future<void> open() async {
    manager.start();
    await pumpEventQueue();
    channel.open();
    await pumpEventQueue();
    _listener = container.listen(chatControllerProvider(_room), (_, _) {});
    await pumpEventQueue();
  }

  /// Drops the socket and lets the manager reconnect on a fresh one.
  Future<void> reconnect() async {
    final before = channels.length;
    channel.serverCloses();
    // The retry timer is 1s; restart() reconnects at once instead.
    manager.restart();
    await pumpEventQueue();
    expect(channels.length, greaterThan(before));
    channel.open();
    await pumpEventQueue();
  }

  List<ChatLine> get lines =>
      container.read(chatControllerProvider(_room)).value ?? const [];

  List<String> get texts => [
    for (final l in lines)
      if (l is ChatMessageLine) l.message.text,
  ];

  void push(Map<String, Object?> data, {String type = 'chat.message'}) =>
      channel.serverSends({
        'op': 'event',
        'topic': 'room:$_room',
        'type': type,
        'data': data,
      });

  Future<void> dispose() async {
    _listener?.close();
    container.dispose();
    await manager.dispose();
  }
}

Future<_Rig> _rig(FakeChatRepository repo) async {
  final rig = _Rig(repo);
  addTearDown(rig.dispose);
  return rig;
}

void main() {
  test('loads the history, then appends live messages in order', () async {
    final rig = await _rig(
      FakeChatRepository(
        histories: [
          okResponse([message('a', 'one'), message('b', 'two')]),
        ],
      ),
    );
    await rig.open();
    expect(rig.texts, ['one', 'two']);

    rig.push(pushJson('c', 'three'));
    await pumpEventQueue();

    expect(rig.texts, ['one', 'two', 'three']);
    expect(rig.channel.opsFor('subscribe'), ['room:$_room']);
  });

  test('incoming chat.message appends once even after reconnect', () async {
    final repo = FakeChatRepository(
      histories: [
        okResponse([message('a', 'one')]),
        // After the reconnect the server also returns a line sent while the
        // socket was down, and the one already shown live.
        okResponse([
          message('a', 'one', at: '2026-11-04T10:00:00Z'),
          message('m', 'missed', at: '2026-11-04T10:00:05Z'),
          message('b', 'live', at: '2026-11-04T10:00:10Z'),
        ]),
      ],
    );
    final rig = await _rig(repo);
    await rig.open();
    rig.push(pushJson('b', 'live', at: 1793786410000));
    await pumpEventQueue();
    expect(rig.texts, ['one', 'live']);

    await rig.reconnect();

    expect(repo.historyCalls, 2);
    expect(rig.channel.opsFor('subscribe'), ['room:$_room']);
    expect(rig.texts, ['one', 'missed', 'live']);

    // The same push again (e.g. redelivery) is still shown once.
    rig.push(pushJson('b', 'live', at: 1793786410000));
    await pumpEventQueue();
    expect(rig.texts, ['one', 'missed', 'live']);
  });

  test('events that arrive while the history loads are kept', () async {
    final hold = Completer<void>();
    final repo = FakeChatRepository(
      histories: [
        okResponse([message('a', 'one')]),
      ],
    )..holdHistory = hold;
    final rig = await _rig(repo);
    await rig.open();

    rig.push(pushJson('b', 'early'));
    rig.push(pushJson('a', 'one'));
    await pumpEventQueue();
    hold.complete();
    await pumpEventQueue();

    expect(rig.texts, ['one', 'early']);
  });

  test('ignores other rooms, malformed payloads and other events', () async {
    final rig = await _rig(FakeChatRepository());
    await rig.open();

    rig.channel.serverSends({
      'op': 'event',
      'topic': 'room:other',
      'type': 'chat.message',
      'data': pushJson('x', 'elsewhere'),
    });
    rig.push({'id': 'y', 'text': 'no sender'});
    rig.push({'userId': 'u-9'}, type: 'notification.created');
    await pumpEventQueue();

    expect(rig.lines, isEmpty);
  });

  test('member.joined and member.left become lines', () async {
    final rig = await _rig(FakeChatRepository());
    await rig.open();

    rig.push({'userId': 'u-5'}, type: 'member.joined');
    rig.push({'userId': 'u-5'}, type: 'member.left');
    await pumpEventQueue();

    final members = rig.lines.whereType<ChatMemberLine>().toList();
    expect(members.map((m) => m.joined), [true, false]);
    expect(members.every((m) => m.userId == 'u-5'), isTrue);
  });

  test(
    'a sent message shows once whether its push or response is first',
    () async {
      final repo = FakeChatRepository()
        ..sendResult = okResponse(message('s1', 'hi'));
      final rig = await _rig(repo);
      await rig.open();
      final controller = rig.container.read(
        chatControllerProvider(_room).notifier,
      );

      // Push first, then the response.
      rig.push(pushJson('s1', 'hi'));
      await pumpEventQueue();
      final res = await controller.send('hi');
      expect(res.success, isTrue);
      expect(rig.texts, ['hi']);

      // Response first, then the push.
      repo.sendResult = okResponse(message('s2', 'again'));
      await controller.send('again');
      rig.push(pushJson('s2', 'again'));
      await pumpEventQueue();
      expect(rig.texts, ['hi', 'again']);
      expect(repo.sent, ['hi', 'again']);
    },
  );

  test('a failed send adds nothing and reports the failure', () async {
    final rig = await _rig(FakeChatRepository());
    await rig.open();
    final controller = rig.container.read(
      chatControllerProvider(_room).notifier,
    );

    final res = await controller.send('lost?');

    expect(res.success, isFalse);
    expect(rig.lines, isEmpty);
  });

  test('notices are local lines', () async {
    final rig = await _rig(FakeChatRepository());
    await rig.open();
    rig.container.read(chatControllerProvider(_room).notifier).addNotice('hi');

    expect(rig.lines.single, isA<ChatNoticeLine>());
  });

  test('a failed history load is an error that a rebuild recovers', () async {
    final repo = FakeChatRepository(
      histories: [
        failedResponse<List<ChatMessage>>(),
        okResponse([message('a', 'one')]),
      ],
    );
    final rig = await _rig(repo);
    await rig.open();
    expect(rig.container.read(chatControllerProvider(_room)).hasError, isTrue);

    rig.container.invalidate(chatControllerProvider(_room));
    await pumpEventQueue();

    expect(rig.texts, ['one']);
  });

  test('leaving the screen unsubscribes and drops the handlers', () async {
    final rig = await _rig(FakeChatRepository());
    await rig.open();
    expect(rig.channel.opsFor('subscribe'), ['room:$_room']);

    rig._listener!.close();
    await pumpEventQueue();
    // autoDispose runs after the listener goes.
    await Future<void>.delayed(Duration.zero);
    expect(rig.channel.opsFor('unsubscribe'), ['room:$_room']);

    // A late push or reconnect must not reach (or revive) the controller.
    final calls = rig.repo.historyCalls;
    rig.push(pushJson('z', 'late'));
    await rig.reconnect();
    expect(rig.repo.historyCalls, calls);
    expect(rig.channel.opsFor('subscribe'), isEmpty);
  });

  group('mergeChatLines', () {
    ChatLine line(String id, String at) =>
        ChatMessageLine(message(id, id, at: at));

    test('keeps only the newest $chatLineLimit lines', () {
      final many = [
        for (var i = 0; i < chatLineLimit + 5; i++)
          ChatNoticeLine(text: '$i', at: DateTime.utc(2026, 1, 1, 0, 0, i)),
      ];
      final merged = mergeChatLines(const [], many);
      expect(merged, hasLength(chatLineLimit));
      expect((merged.first as ChatNoticeLine).text, '5');
    });

    test('orders late-arriving older messages by time', () {
      final merged = mergeChatLines(
        [line('a', '2026-11-04T10:00:00Z'), line('c', '2026-11-04T10:00:10Z')],
        [line('b', '2026-11-04T10:00:05Z')],
      );
      expect(
        [for (final l in merged) (l as ChatMessageLine).message.id],
        ['a', 'b', 'c'],
      );
    });
  });
}
