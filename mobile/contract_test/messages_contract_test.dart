import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/data/message_repository.dart';
import 'package:letslive_mobile/features/messages/domain/dm_event.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

import 'harness.dart';

const Duration _pushWait = Duration(seconds: 10);
// There is no subscribe acknowledgement; give the gateway a moment to
// register the socket before the first message is sent.
const Duration _socketSettle = Duration(seconds: 1);

/// Direct messages against the real dev stack with two clients: A writes to
/// B, and B, connected over a real `RealtimeManager`, gets the pushes. The
/// push, the send response and the history must carry the same message id,
/// because the app merges them by that id.
void main() {
  ContractContext? ctx;
  RealtimeManager? bSocket;
  late MessageRepository a;
  late MessageRepository b;
  late String aId;
  late String bId;
  late Conversation conversation;

  Future<String> idOf(ContractUser user) async {
    final me = await user.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    return me.data!.id;
  }

  setUpAll(() async {
    final context = ctx = await startContractContext();
    a = MessageRepository(context.userA.client);
    b = MessageRepository(context.userB.client);
    aId = await idOf(context.userA);
    bId = await idOf(context.userB);
  });

  tearDownAll(() async {
    try {
      await bSocket?.dispose();
    } finally {
      await ctx?.dispose();
    }
  });

  test('a user cannot start a conversation with themself', () async {
    final res = await a.create(
      type: ConversationType.dm,
      participantIds: [aId],
    );

    expect(res.success, isFalse);
  });

  test(
    'A starts a DM with B; starting it again answers the same one',
    () async {
      final first = await a.create(
        type: ConversationType.dm,
        participantIds: [bId],
      );
      expect(first.success, isTrue, reason: '${first.key}: ${first.message}');
      conversation = first.data!;
      expect(conversation.type, ConversationType.dm);
      expect(conversation.participants.map((p) => p.userId).toSet(), {
        aId,
        bId,
      });
      expect(conversation.lastMessage, isNull);

      final again = await a.create(
        type: ConversationType.dm,
        participantIds: [bId],
      );
      expect(again.data!.id, conversation.id);

      final listed = await b.list();
      expect(listed.success, isTrue, reason: listed.key);
      expect(listed.data!.map((c) => c.id), contains(conversation.id));
      final detail = await b.byId(conversation.id);
      expect(detail.data!.id, conversation.id);
    },
  );

  test(
    'a stranger gets a not-found style error for the conversation',
    () async {
      final res = await a.byId('000000000000000000000000');

      expect(res.success, isFalse);
      expect(res.statusCode, anyOf(403, 404));
    },
  );

  test(
    'a message A sends reaches connected B once, with the same id everywhere',
    () async {
      final context = ctx!;
      final socket = bSocket = RealtimeManager(
        url: context.config.wsBaseUrl,
        cookieJar: context.userB.cookieJar,
        refreshSession: context.userB.client.refreshSession,
      );
      final pushes = <DmNewMessage>[];
      final typing = Completer<DmTyping>();
      final firstPush = Completer<void>();
      socket
        ..onEvent(DmEventType.newMessage, (data) {
          final event = parseDmEvent(DmEventType.newMessage, data);
          if (event is! DmNewMessage) return;
          pushes.add(event);
          if (!firstPush.isCompleted) firstPush.complete();
        })
        ..onEvent(DmEventType.userTyping, (data) {
          final event = parseDmEvent(DmEventType.userTyping, data);
          if (event is DmTyping && !typing.isCompleted) typing.complete(event);
        });
      final connected = socket.status.firstWhere(
        (s) => s == RealtimeStatus.connected,
      );
      socket.start();
      await connected.timeout(_pushWait);
      await Future<void>.delayed(_socketSettle);

      const text = 'Xin chào 🎉 :fire: tin nhắn trực tiếp';
      final sent = await a.send(conversation.id, text);
      expect(sent.success, isTrue, reason: '${sent.key}: ${sent.message}');
      await firstPush.future.timeout(_pushWait);

      expect(pushes, hasLength(1));
      expect(pushes.single.conversationId, conversation.id);
      expect(pushes.single.message.id, sent.data!.id);
      expect(pushes.single.message.text, text);
      expect(pushes.single.message.senderId, aId);

      final history = await b.messages(conversation.id);
      expect(history.data!.map((m) => m.id), [sent.data!.id]);

      // B has one unread message until B reads the conversation.
      final unread = await b.unreadCounts();
      expect(unread.data, {conversation.id: 1});
      final read = await b.markRead(conversation.id);
      expect(read.success, isTrue, reason: read.key);
      expect((await b.unreadCounts()).data ?? const {}, isEmpty);

      // The list preview follows the newest message.
      final listed = await b.list();
      final item = listed.data!.firstWhere((c) => c.id == conversation.id);
      expect(item.lastMessage?.text, text);

      // A's typing hint reaches B with A's username.
      final hint = await a.typing(conversation.id, isTyping: true);
      expect(hint.success, isTrue, reason: hint.key);
      final event = await typing.future.timeout(_pushWait);
      expect(event.conversationId, conversation.id);
      expect(event.username, context.userA.username);
    },
  );

  test('messages page back by cursor without overlap', () async {
    final ids = <String>[];
    for (var i = 0; i < 4; i++) {
      final sent = await a.send(conversation.id, 'page test $i');
      ids.add(sent.data!.id);
    }

    final newest = await b.messages(conversation.id, limit: 2);
    expect(newest.data!.map((m) => m.id), ids.sublist(2));
    final older = await b.messages(
      conversation.id,
      before: newest.data!.first.id,
      limit: 2,
    );
    expect(older.data!.map((m) => m.id), ids.sublist(0, 2));
  });

  test('only the sender can edit or delete a message', () async {
    final sent = await a.send(conversation.id, 'original');
    final id = sent.data!.id;

    final foreignEdit = await b.edit(conversation.id, id, 'hijacked');
    expect(foreignEdit.success, isFalse);
    expect(foreignEdit.statusCode, 403);
    final foreignDelete = await b.delete(conversation.id, id);
    expect(foreignDelete.success, isFalse);
    expect(foreignDelete.statusCode, 403);

    final edited = await a.edit(conversation.id, id, 'edited');
    expect(edited.success, isTrue, reason: '${edited.key}: ${edited.message}');
    expect(edited.data!.text, 'edited');
    expect(edited.data!.isEdited, isTrue);

    final deleted = await a.delete(conversation.id, id);
    expect(deleted.success, isTrue, reason: deleted.key);
    final history = await b.messages(conversation.id, limit: 1);
    expect(history.data!.single.isDeleted, isTrue);
    expect(history.data!.single.text, isEmpty);
  });

  test('a group can be created and left', () async {
    final group = await a.create(
      type: ConversationType.group,
      participantIds: [bId],
      name: 'Contract crew',
    );
    expect(group.success, isTrue, reason: '${group.key}: ${group.message}');
    expect(group.data!.type, ConversationType.group);
    expect(group.data!.name, 'Contract crew');

    final left = await b.leave(group.data!.id);
    expect(left.success, isTrue, reason: '${left.key}: ${left.message}');
    final after = await b.byId(group.data!.id);
    expect(after.success, isFalse);
  });
}
