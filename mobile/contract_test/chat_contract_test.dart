import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/realtime/realtime_frame.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_repository.dart';

import 'harness.dart';

const Duration _pushWait = Duration(seconds: 10);
// There is no subscribe acknowledgement; give the gateway a moment to
// register the topic before the first message is sent.
const Duration _subscribeSettle = Duration(seconds: 1);

/// Live chat against the real dev stack: client A says something in streamer
/// S's room and client B, subscribed over a real `RealtimeManager`, gets the
/// push. The push, the send response and the history must carry the same
/// message id, because the app merges them by that id.
void main() {
  ContractContext? ctx;
  RealtimeManager? bSocket;
  late String roomId;
  late ChatRepository aChat;
  late ChatRepository bChat;

  setUpAll(() async {
    final context = ctx = await startContractContext();
    aChat = ChatRepository(context.userA.client);
    bChat = ChatRepository(context.userB.client);
    final me = await context.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    // The room id is the streamer's user id; A plays the streamer.
    roomId = me.data!.id;
  });

  tearDownAll(() async {
    try {
      await bSocket?.dispose();
    } finally {
      await ctx?.dispose();
    }
  });

  test(
    'GET /chat-commands?roomId= answers an empty list for a new room',
    () async {
      final res = await bChat.commandsFor(roomId);

      expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
      expect(res.data ?? const [], isEmpty);
    },
  );

  test('GET /messages?roomId= is empty for a room nobody spoke in', () async {
    final res = await bChat.history(roomId);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data ?? const [], isEmpty);
  });

  test(
    'a message A sends reaches subscribed B once, with the same id everywhere',
    () async {
      final context = ctx!;
      final socket = bSocket = RealtimeManager(
        url: context.config.wsBaseUrl,
        cookieJar: context.userB.cookieJar,
        refreshSession: context.userB.client.refreshSession,
      );
      final topic = realtimeRoomTopic(roomId);
      final pushes = <ChatMessage>[];
      final firstPush = Completer<void>();
      socket.onEvent(RealtimeEvent.chatMessage, (data) {
        if (data is! Map<String, Object?>) return;
        pushes.add(ChatMessage.fromJson(data));
        if (!firstPush.isCompleted) firstPush.complete();
      }, topic: topic);
      final connected = socket.status.firstWhere(
        (s) => s == RealtimeStatus.connected,
      );
      socket
        ..subscribe(topic)
        ..start();
      await connected.timeout(_pushWait);
      await Future<void>.delayed(_subscribeSettle);

      const text = 'Xin chào, tin nhắn thử 🎉 :fire:';
      final sent = await aChat.send(roomId, text);
      expect(sent.success, isTrue, reason: '${sent.key}: ${sent.message}');
      await firstPush.future.timeout(_pushWait);

      expect(pushes, hasLength(1));
      expect(pushes.single.text, text);
      expect(pushes.single.userId, sent.data!.userId);
      expect(pushes.single.id, sent.data!.id);

      final history = await bChat.history(roomId);
      expect(history.success, isTrue, reason: history.key);
      expect(history.data!.map((m) => m.id), [sent.data!.id]);
      expect(history.data!.single.text, text);
    },
  );

  test('POST /messages rejects an empty text', () async {
    final res = await aChat.send(roomId, '   ');

    expect(res.success, isFalse);
    expect(res.statusCode, 400);
  });
}
