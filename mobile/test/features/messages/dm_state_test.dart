import 'package:fake_async/fake_async.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';

import 'message_fakes.dart';

List<String> _ids(DmRig rig, [String id = 'c-1']) => [
  for (final m
      in rig.container.read(dmThreadProvider(id)).requireValue.messages)
    m.id,
];

void main() {
  test('messages page by cursor without duplicates', () async {
    final repo = FakeMessageRepository()
      ..threads['c-1'] = [for (var i = 0; i < 120; i++) numbered(i)];
    final rig = newRig(repo)..keepAlive(dmThreadProvider('c-1'));
    await rig.connect();
    await rig.container.read(dmThreadProvider('c-1').future);

    var thread = rig.container.read(dmThreadProvider('c-1')).requireValue;
    expect(thread.messages, hasLength(dmMessagesPageSize));
    expect(thread.messages.first.id, 'm0070');
    expect(thread.hasMore, isTrue);

    // A push of a message already loaded must not show twice.
    rig.push('dm:new_message', {
      'conversationId': 'c-1',
      'message': messageJson(
        'm0119',
        'text 119',
        at: '2026-11-04T10:01:59.000Z',
      ),
    });
    await pumpEventQueue();

    final notifier = rig.container.read(dmThreadProvider('c-1').notifier);
    await notifier.loadOlder();
    thread = rig.container.read(dmThreadProvider('c-1')).requireValue;
    expect(repo.cursors, [null, 'm0070']);
    expect(thread.messages, hasLength(100));
    expect(thread.hasMore, isTrue);

    await notifier.loadOlder();
    thread = rig.container.read(dmThreadProvider('c-1')).requireValue;
    expect(repo.cursors.last, 'm0020');
    expect(thread.hasMore, isFalse);
    final ids = _ids(rig);
    expect(ids.toSet(), hasLength(120));
    expect(ids, [
      for (var i = 0; i < 120; i++) 'm${i.toString().padLeft(4, '0')}',
    ]);
  });

  test('reconnect refetches the open conversation and merges without '
      'duplicates', () async {
    final repo = FakeMessageRepository()
      ..threads['c-1'] = [numbered(1), numbered(2)];
    final rig = newRig(repo)..keepAlive(dmThreadProvider('c-1'));
    await rig.connect();
    await rig.container.read(dmThreadProvider('c-1').future);
    rig.push('dm:new_message', {
      'conversationId': 'c-1',
      'message': messageJson('m0003', 'text 3', at: '2026-11-04T10:00:03.000Z'),
    });
    await pumpEventQueue();
    expect(_ids(rig), ['m0001', 'm0002', 'm0003']);

    // While the socket is down m0003 is also saved server side, and m0004
    // is sent; both come back with the refetch.
    repo.threads['c-1'] = [numbered(1), numbered(2), numbered(3), numbered(4)];
    await rig.reconnect();

    expect(repo.messageCalls, 2);
    expect(_ids(rig), ['m0001', 'm0002', 'm0003', 'm0004']);
    // Marked read on open, for the pushed message and after the refetch.
    expect(repo.markedRead, ['c-1', 'c-1', 'c-1']);
  });

  test(
    'a reconnect gap wider than a page replaces the loaded messages',
    () async {
      final repo = FakeMessageRepository()
        ..threads['c-1'] = [numbered(1), numbered(2)];
      final rig = newRig(repo)..keepAlive(dmThreadProvider('c-1'));
      await rig.connect();
      await rig.container.read(dmThreadProvider('c-1').future);

      repo.threads['c-1'] = [for (var i = 10; i < 200; i++) numbered(i)];
      await rig.reconnect();

      final thread = rig.container.read(dmThreadProvider('c-1')).requireValue;
      expect(thread.messages, hasLength(dmMessagesPageSize));
      expect(thread.messages.first.id, 'm0150');
      expect(thread.hasMore, isTrue);
    },
  );

  test('opening a conversation marks it read and decrements badge', () async {
    final repo = FakeMessageRepository()
      ..threads['c-1'] = [numbered(1)]
      ..unread.addAll({'c-1': 2, 'c-2': 1});
    final rig = newRig(repo)..keepAlive(unreadDmCountProvider);
    await rig.connect();
    await rig.container.read(dmUnreadProvider.future);
    expect(rig.container.read(unreadDmCountProvider), 3);

    rig.keepAlive(dmThreadProvider('c-1'));
    await rig.container.read(dmThreadProvider('c-1').future);
    await pumpEventQueue();

    expect(repo.markedRead, ['c-1']);
    expect(rig.container.read(unreadDmCountProvider), 1);
  });

  test('a push counts as unread only in a conversation that is not open, '
      'and never when it is your own', () async {
    final repo = FakeMessageRepository()..threads['c-1'] = [numbered(1)];
    final rig = newRig(repo)
      ..keepAlive(unreadDmCountProvider)
      ..keepAlive(dmThreadProvider('c-1'));
    await rig.connect();
    await rig.container.read(dmUnreadProvider.future);
    await rig.container.read(dmThreadProvider('c-1').future);

    void push(String conversation, String id, String sender) =>
        rig.push('dm:new_message', {
          'conversationId': conversation,
          'message': messageJson(
            id,
            'hi',
            conversationId: conversation,
            senderId: sender,
          ),
        });
    push('c-1', 'x1', 'u-2'); // open
    push('c-2', 'x2', me); // own echo
    await pumpEventQueue();
    expect(rig.container.read(unreadDmCountProvider), 0);

    push('c-2', 'x3', 'u-2');
    await pumpEventQueue();
    expect(rig.container.read(unreadDmCountProvider), 1);
  });

  test('typing indicator expires after the same timeout as web', () {
    expect(dmTypingTimeout, const Duration(seconds: 5));
    fakeAsync((async) {
      final repo = FakeMessageRepository();
      final rig = DmRig(repo);
      rig.keepAlive(dmPresenceProvider);
      rig.manager.start();
      async.flushMicrotasks();
      rig.channel.open();
      async.flushMicrotasks();

      List<String> typing() =>
          rig.container.read(dmPresenceProvider).typing['c-1'] ?? const [];
      void push(String type) {
        rig.push(type, {
          'conversationId': 'c-1',
          'userId': 'u-2',
          'username': 'bob',
        });
        async.flushMicrotasks();
      }

      push('dm:user_typing');
      expect(typing(), ['bob']);
      async.elapse(const Duration(milliseconds: 4999));
      expect(typing(), ['bob']);
      async.elapse(const Duration(milliseconds: 1));
      expect(typing(), isEmpty);

      // Typing again restarts the timer; "stopped" clears at once.
      push('dm:user_typing');
      async.elapse(const Duration(seconds: 3));
      push('dm:user_typing');
      async.elapse(const Duration(seconds: 3));
      expect(typing(), ['bob']);
      push('dm:user_stopped_typing');
      expect(typing(), isEmpty);

      rig.dispose();
      async.flushMicrotasks();
    });
  });

  test('presence events mark users online and offline', () async {
    final rig = newRig(FakeMessageRepository())..keepAlive(dmPresenceProvider);
    await rig.connect();

    rig.push('dm:user_online', {'userId': 'u-2'});
    await pumpEventQueue();
    expect(rig.container.read(dmPresenceProvider).online, {'u-2'});

    rig.push('dm:user_offline', {'userId': 'u-2'});
    await pumpEventQueue();
    expect(rig.container.read(dmPresenceProvider).online, isEmpty);
  });

  test('edit and delete only work on own messages', () async {
    final own = message('a', 'mine', senderId: me);
    final other = message('b', 'theirs');
    final image = message('c', 'Sent an image', senderId: me, type: 'image');
    final system = message('d', 'joined', senderId: me, type: 'system');
    final gone = message('e', '', senderId: me, isDeleted: true);

    expect(canEditMessage(own, me), isTrue);
    expect(canDeleteMessage(own, me), isTrue);
    expect(canEditMessage(other, me), isFalse);
    expect(canDeleteMessage(other, me), isFalse);
    expect(canEditMessage(image, me), isFalse);
    expect(canDeleteMessage(image, me), isTrue);
    expect(canEditMessage(system, me), isFalse);
    expect(canDeleteMessage(system, me), isFalse);
    expect(canEditMessage(gone, me), isFalse);
    expect(canDeleteMessage(gone, me), isFalse);

    final repo = FakeMessageRepository()..threads['c-1'] = [own, other];
    final rig = newRig(repo)..keepAlive(dmThreadProvider('c-1'));
    await rig.container.read(dmThreadProvider('c-1').future);
    final notifier = rig.container.read(dmThreadProvider('c-1').notifier);

    expect(await notifier.edit('b', 'changed'), isNull);
    expect(await notifier.delete('b'), isNull);
    expect(repo.threads['c-1']![1].text, 'theirs');

    expect((await notifier.edit('a', 'fixed'))?.success, isTrue);
    expect(
      rig.container
          .read(dmThreadProvider('c-1'))
          .requireValue
          .messages
          .first
          .text,
      'fixed',
    );
    expect((await notifier.delete('a'))?.success, isTrue);
    expect(
      rig.container
          .read(dmThreadProvider('c-1'))
          .requireValue
          .messages
          .first
          .isDeleted,
      isTrue,
    );
  });

  test('mergeDmMessages drops duplicates and orders by time', () {
    final a = numbered(1);
    final b = numbered(2);
    final c = numbered(3);
    final merged = mergeDmMessages([b], [c, a, b]);
    expect([for (final m in merged) m.id], ['m0001', 'm0002', 'm0003']);
  });

  test('conversation model keeps update fields and last message', () {
    final c = conversation('c-1');
    final updated = c.withUpdate({'name': 'Team', 'type': 'dm'});
    expect(updated.name, 'Team');
    expect(updated.type, ConversationType.dm);
  });
}
