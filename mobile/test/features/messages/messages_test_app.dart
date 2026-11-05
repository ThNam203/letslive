import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/realtime/realtime_manager.dart';
import 'package:letslive_mobile/core/router/app_router.dart';

import '../../core/network/fake_adapter.dart';
import '../../core/realtime/fake_channel.dart';
import '../auth/auth_test_app.dart';
import '../home/home_test_app.dart';
import '../search/search_test_app.dart';
import 'message_fakes.dart';

/// [FakeSearchServer] plus the `/conversations` endpoints, over in-memory
/// [conversations] and [threads] (oldest first). What the app sends is
/// recorded; the `*Status` fields make a family of endpoints fail.
class FakeMessagesServer extends FakeSearchServer {
  final List<Map<String, Object?>> conversations = [];
  final Map<String, List<Map<String, Object?>>> threads = {};
  final Map<String, int> unread = {};

  /// Status of `GET /conversations` (e.g. 503).
  int listStatus = 200;

  /// Status of `POST .../messages` (e.g. 503).
  int sendStatus = 200;

  /// Bodies of `POST .../messages`, in order.
  final List<String> sentTexts = [];

  /// Bodies of `POST .../messages` that carried pictures, in order.
  final List<Map<String, Object?>> pictureBodies = [];

  /// Number of `POST /upload-file` calls, and the 0-based call numbers that
  /// answer 503 instead of a stored path.
  int uploadCalls = 0;
  final Set<int> failingUploads = {};

  /// Conversation ids `POST .../read` was called for.
  final List<String> readCalls = [];

  /// Bodies of `POST /conversations`.
  final List<Map<String, Object?>> created = [];

  /// Typing hints posted: conversation id and state.
  final List<(String, String)> typingPosts = [];

  ResponseBody _fail(int status, String key) => jsonBody({
    'requestId': 'req-err',
    'success': false,
    'code': 1,
    'key': key,
    'message': 'Nope',
  }, status: status);

  Map<String, Object?>? _conversation(String id) {
    for (final c in conversations) {
      if (c['_id'] == id) return c;
    }
    return null;
  }

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    if (o.path == ApiEndpoints.uploadFile && o.method == 'POST') {
      final call = uploadCalls++;
      return failingUploads.contains(call)
          ? _fail(503, 'res_err_internal_server')
          : okEnvelope('/files/uploads/pic-$call.jpg');
    }
    if (!o.path.startsWith(ApiEndpoints.conversations)) {
      return super.handle(o);
    }
    final parts = o.path.split('/'); // ['', 'conversations', id?, sub?, ...]
    final body = (o.data as Map<String, Object?>?) ?? const {};
    if (parts.length == 2) {
      if (o.method == 'GET') {
        if (listStatus != 200) return _fail(listStatus, 'res_err_internal');
        final page = o.queryParameters['page'] as int;
        final limit = o.queryParameters['limit'] as int;
        return okEnvelope(
          conversations.skip(page * limit).take(limit).toList(),
        );
      }
      created.add(body);
      final others = (body['participantIds'] as List).cast<String>();
      final made = conversationJson(
        'new-1',
        type: body['type'] as String,
        name: body['name'] as String?,
        otherId: others.first,
        otherName: 'picked',
      );
      conversations.insert(0, made);
      threads['new-1'] = [];
      return okEnvelope(made);
    }
    if (parts[2] == 'unread-counts') return okEnvelope(Map.of(unread));
    final id = parts[2];
    final conversation = _conversation(id);
    if (conversation == null) {
      return _fail(404, 'res_err_conversation_not_found');
    }
    if (parts.length == 3) return okEnvelope(conversation);
    final thread = threads[id] ??= [];
    switch ((o.method, parts[3])) {
      case ('GET', 'messages'):
        final before = o.queryParameters['before'] as String?;
        final end = before == null
            ? thread.length
            : thread.indexWhere((m) => m['_id'] == before);
        final older = thread.sublist(0, end);
        final limit = o.queryParameters['limit'] as int;
        return okEnvelope(
          older.skip(older.length > limit ? older.length - limit : 0).toList(),
        );
      case ('POST', 'messages'):
        final text = body['text'] as String;
        sentTexts.add(text);
        if (sendStatus != 200) return _fail(sendStatus, 'res_err_internal');
        final pictures = (body['imageUrls'] as List?)?.cast<String>() ?? [];
        if (pictures.isNotEmpty) pictureBodies.add(body);
        final sent = messageJson(
          's-${sentTexts.length}',
          text,
          type: body['type'] as String,
          imageUrls: pictures,
          conversationId: id,
          senderId: me,
          senderUsername: 'abcdef',
          at: DateTime.now().toUtc().toIso8601String(),
        );
        thread.add(sent);
        return okEnvelope(sent);
      case ('PATCH', 'messages'):
        final index = thread.indexWhere((m) => m['_id'] == parts[4]);
        thread[index] = {
          ...thread[index],
          'text': body['text'],
          'updatedAt': '2026-12-01T00:00:00.000Z',
        };
        return okEnvelope(thread[index]);
      case ('DELETE', 'messages'):
        final index = thread.indexWhere((m) => m['_id'] == parts[4]);
        thread[index] = {...thread[index], 'text': '', 'isDeleted': true};
        return okEnvelope(null);
      case ('POST', 'read'):
        readCalls.add(id);
        unread.remove(id);
        return okEnvelope(null);
      case ('POST', 'typing'):
        typingPosts.add((id, body['state'] as String));
        return okEnvelope(null);
    }
    return _fail(404, 'res_err_route_not_found');
  }
}

/// A realtime socket the test controls, installed with [overrides].
class FakeSocket {
  FakeSocket() {
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
  }

  late final RealtimeManager manager;
  final List<FakeChannel> channels = [];

  List<Override> get overrides => [
    realtimeManagerProvider.overrideWithValue(manager),
  ];

  /// Connects the socket; call after the app is pumped.
  Future<void> connect(WidgetTester tester) async {
    manager.start();
    await tester.pump();
    channels.last.open();
    await tester.pump();
  }

  /// Delivers a DM push.
  Future<void> push(
    WidgetTester tester,
    String type,
    Map<String, Object?> data,
  ) async {
    channels.last.serverSends({
      'op': 'event',
      'topic': 'user:$me',
      'type': type,
      'data': data,
    });
    await tester.pump();
    await tester.pump();
  }

  Future<void> dispose() => manager.dispose();
}

/// Pumps the signed-in app against [server] on the Messages tab.
Future<TestApp> pumpMessages(
  WidgetTester tester,
  FakeMessagesServer server, {
  FakeSocket? socket,
  double width = 800,
  double height = 1200,
  double textScale = 1,
  List<Override> overrides = const [],
}) async {
  final app = await pumpHome(
    tester,
    server,
    width: width,
    height: height,
    textScale: textScale,
    overrides: [...?socket?.overrides, ...overrides],
  );
  await tester.tap(
    find.descendant(
      of: find.byType(NavigationBar),
      matching: find.text('Messages'),
    ),
  );
  await tester.pumpAndSettle();
  return app;
}

/// Opens the conversation [id] the way a tap does (pushed above the tabs).
Future<void> openConversation(
  WidgetTester tester,
  TestApp app,
  String id,
) async {
  unawaited(
    app.container.read(appRouterProvider).push(AppRoutes.conversation(id)),
  );
  await tester.pumpAndSettle();
}

/// A thread of [count] numbered messages from `u-2` for conversation [id].
List<Map<String, Object?>> threadJson(String id, int count) => [
  for (var i = 0; i < count; i++)
    messageJson(
      'm${i.toString().padLeft(4, '0')}',
      'text $i',
      conversationId: id,
      at: DateTime.utc(
        2026,
        11,
        4,
        10,
      ).add(Duration(seconds: i)).toIso8601String(),
    ),
];
