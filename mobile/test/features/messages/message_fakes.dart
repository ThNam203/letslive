import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart' show ProviderListenable;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/data/message_repository.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';

import '../../core/realtime/fake_channel.dart';

/// The signed-in user of the unit tests.
const String me = 'u-1';

ApiResponse<T> ok<T>(T data) => ApiResponse<T>(
  success: true,
  statusCode: 200,
  code: 0,
  key: '',
  message: '',
  requestId: 'r',
  data: data,
);

ApiResponse<T> failed<T>({int status = 503, String key = 'res_err_internal'}) =>
    ApiResponse<T>(
      success: false,
      statusCode: status,
      code: 1,
      key: key,
      message: 'Server down',
      requestId: 'r',
    );

/// A message as the backend serializes it (`dto.DmMessageResponse`).
Map<String, Object?> messageJson(
  String id,
  String text, {
  String conversationId = 'c-1',
  String senderId = 'u-2',
  String senderUsername = 'bob',
  String type = 'text',
  bool isDeleted = false,
  String at = '2026-11-04T10:00:00.000Z',
  String? updatedAt,
  List<String> imageUrls = const [],
}) => {
  '_id': id,
  'conversationId': conversationId,
  'senderId': senderId,
  'senderUsername': senderUsername,
  'type': type,
  'text': text,
  'imageUrls': imageUrls,
  'replyTo': null,
  'isDeleted': isDeleted,
  'readBy': <Object?>[],
  'createdAt': at,
  'updatedAt': updatedAt ?? at,
  '__v': 0,
};

DmMessage message(
  String id,
  String text, {
  String conversationId = 'c-1',
  String senderId = 'u-2',
  String type = 'text',
  bool isDeleted = false,
  DateTime? at,
}) => DmMessage.fromJson(
  messageJson(
    id,
    text,
    conversationId: conversationId,
    senderId: senderId,
    senderUsername: senderId == me ? 'abcdef' : 'bob',
    type: type,
    isDeleted: isDeleted,
    at: (at ?? DateTime.utc(2026, 11, 4, 10)).toUtc().toIso8601String(),
  ),
);

/// Message number [i] of a long thread: one second after the previous one.
DmMessage numbered(int i, {String conversationId = 'c-1'}) => message(
  'm${i.toString().padLeft(4, '0')}',
  'text $i',
  conversationId: conversationId,
  at: DateTime.utc(2026, 11, 4, 10).add(Duration(seconds: i)),
);

/// A conversation as the backend serializes it.
Map<String, Object?> conversationJson(
  String id, {
  String type = 'dm',
  String? name,
  String otherId = 'u-2',
  String otherName = 'bob',
  Map<String, Object?>? lastMessage,
  String updatedAt = '2026-11-04T10:00:00.000Z',
}) => {
  '_id': id,
  'type': type,
  'name': name,
  'avatarUrl': null,
  'createdBy': me,
  'participants': [
    {
      'userId': me,
      'username': 'abcdef',
      'profilePicture': null,
      'role': 'owner',
      'joinedAt': '2026-11-04T09:00:00.000Z',
      'lastReadMessageId': null,
      'isMuted': false,
    },
    {
      'userId': otherId,
      'username': otherName,
      'profilePicture': null,
      'role': 'member',
      'joinedAt': '2026-11-04T09:00:00.000Z',
      'lastReadMessageId': null,
      'isMuted': false,
    },
  ],
  'lastMessage': lastMessage,
  'createdAt': '2026-11-04T09:00:00.000Z',
  'updatedAt': updatedAt,
  '__v': 0,
};

Conversation conversation(String id, {String type = 'dm', String? name}) =>
    Conversation.fromJson(conversationJson(id, type: type, name: name));

/// A scriptable server in memory: [threads] hold each conversation's
/// messages oldest first and `messages` answers cursor pages from them, like
/// the real endpoint. Everything the app sends is recorded.
class FakeMessageRepository implements MessageRepository {
  final Map<String, List<DmMessage>> threads = {};
  final Map<String, int> unread = {};
  final List<Conversation> conversations = [];

  /// Cursors (`before`) the app asked messages with, per call.
  final List<String?> cursors = [];
  final List<String> markedRead = [];
  final List<String> sent = [];

  /// The picture paths sent with each message in [sent] (empty for text).
  final List<List<String>> sentImageUrls = [];
  final List<(String, bool)> typingSent = [];
  int messageCalls = 0;
  int unreadCalls = 0;

  /// When set, `unreadCounts` waits for it (its answer then lands late).
  Completer<void>? holdUnread;

  /// When set, `messages` waits for it.
  Completer<void>? holdMessages;

  /// Answers every `messages` call with this failure when set.
  ApiResponse<List<DmMessage>>? messagesFailure;

  @override
  Future<ApiResponse<List<DmMessage>>> messages(
    String conversationId, {
    String? before,
    int limit = dmMessagesPageSize,
  }) async {
    messageCalls++;
    cursors.add(before);
    await holdMessages?.future;
    final failure = messagesFailure;
    if (failure != null) return failure;
    final all = threads[conversationId] ?? const <DmMessage>[];
    final end = before == null
        ? all.length
        : all.indexWhere((m) => m.id == before);
    final older = all.sublist(0, end < 0 ? all.length : end);
    return ok(
      older.skip(older.length > limit ? older.length - limit : 0).toList(),
    );
  }

  @override
  Future<ApiResponse<DmMessage>> send(
    String conversationId,
    String text, {
    List<String> imageUrls = const [],
  }) async {
    sent.add(text);
    sentImageUrls.add(imageUrls);
    final created = message(
      's-${sent.length}',
      text,
      conversationId: conversationId,
      senderId: me,
      at: DateTime.utc(2026, 11, 4, 12).add(Duration(minutes: sent.length)),
    );
    (threads[conversationId] ??= []).add(created);
    return ok(created);
  }

  @override
  Future<ApiResponse<DmMessage>> edit(
    String conversationId,
    String messageId,
    String text,
  ) async {
    final thread = threads[conversationId]!;
    final index = thread.indexWhere((m) => m.id == messageId);
    final edited = thread[index].withText(text, DateTime.utc(2026, 11, 5));
    thread[index] = edited;
    return ok(edited);
  }

  @override
  Future<ApiResponse<Object?>> delete(
    String conversationId,
    String messageId,
  ) async {
    final thread = threads[conversationId]!;
    final index = thread.indexWhere((m) => m.id == messageId);
    thread[index] = thread[index].asDeleted();
    return ok<Object?>(null);
  }

  @override
  Future<ApiResponse<Object?>> markRead(
    String conversationId, {
    String? messageId,
  }) async {
    markedRead.add(conversationId);
    unread.remove(conversationId);
    return ok<Object?>(null);
  }

  @override
  Future<ApiResponse<Object?>> typing(
    String conversationId, {
    required bool isTyping,
  }) async {
    typingSent.add((conversationId, isTyping));
    return ok<Object?>(null);
  }

  @override
  Future<ApiResponse<Map<String, int>>> unreadCounts() async {
    unreadCalls++;
    final snapshot = Map.of(unread);
    await holdUnread?.future;
    return ok(snapshot);
  }

  @override
  Future<ApiResponse<List<Conversation>>> list({
    int page = 0,
    int limit = dmConversationsPageSize,
  }) async => ok(conversations.skip(page * limit).take(limit).toList());

  @override
  Future<ApiResponse<Conversation>> byId(String id) async {
    final found = conversations.where((c) => c.id == id);
    return found.isEmpty
        ? failed(status: 404, key: 'res_err_conversation_not_found')
        : ok(found.first);
  }

  @override
  Future<ApiResponse<Conversation>> create({
    required ConversationType type,
    required List<String> participantIds,
    String? name,
  }) => throw UnimplementedError('not used by the unit tests');

  @override
  Future<ApiResponse<Conversation>> update(
    String id, {
    String? name,
    String? avatarUrl,
  }) => throw UnimplementedError('not used by the unit tests');

  @override
  Future<ApiResponse<Object?>> leave(String id) =>
      throw UnimplementedError('not used by the unit tests');

  @override
  Future<ApiResponse<Conversation>> addParticipant(
    String id, {
    required String userId,
    required String username,
    String? profilePicture,
  }) => throw UnimplementedError('not used by the unit tests');

  @override
  Future<ApiResponse<Conversation>> removeParticipant(
    String id,
    String userId,
  ) => throw UnimplementedError('not used by the unit tests');
}

/// A real [RealtimeManager] over [FakeChannel]s plus the provider container
/// the DM providers live in, with [repo] as the server.
class DmRig {
  DmRig(this.repo) {
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
        messageRepositoryProvider.overrideWithValue(repo),
        currentUserIdProvider.overrideWithValue(me),
      ],
    );
  }

  final FakeMessageRepository repo;
  late final RealtimeManager manager;
  late final ProviderContainer container;
  final List<FakeChannel> channels = [];
  final List<ProviderSubscription<Object?>> _listeners = [];

  FakeChannel get channel => channels.last;

  /// Keeps [provider] alive, like a screen watching it.
  void keepAlive(ProviderListenable<Object?> provider) =>
      _listeners.add(container.listen(provider, (_, _) {}));

  /// Connects the socket (handshake done).
  Future<void> connect() async {
    manager.start();
    await pumpEventQueue();
    channel.open();
    await pumpEventQueue();
  }

  /// Drops the socket and reconnects on a fresh one.
  Future<void> reconnect() async {
    final before = channels.length;
    channel.serverCloses();
    manager.restart();
    await pumpEventQueue();
    expect(channels.length, greaterThan(before));
    channel.open();
    await pumpEventQueue();
  }

  /// Delivers a DM push (they arrive on the user's own topic).
  void push(String type, Map<String, Object?> data) => channel.serverSends({
    'op': 'event',
    'topic': 'user:$me',
    'type': type,
    'data': data,
  });

  Future<void> dispose() async {
    for (final listener in _listeners) {
      listener.close();
    }
    container.dispose();
    await manager.dispose();
  }
}

DmRig newRig(FakeMessageRepository repo) {
  final rig = DmRig(repo);
  addTearDown(rig.dispose);
  return rig;
}
