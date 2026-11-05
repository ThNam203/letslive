import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../home/home_test_app.dart';
import '../search/search_test_app.dart';
import 'fake_player.dart';

/// [FakeSearchServer] plus `GET /livestreams?userId=`: [liveByUser] holds
/// the live stream of each user (absent = not live, answered as a success
/// with no data, like the backend). Unknown users on `/user/<id>` answer
/// the backend's 404 `res_err_user_not_found`.
class FakeWatchServer extends FakeSearchServer {
  Map<String, Map<String, Object?>> liveByUser = {};

  /// Status `/livestreams?userId=` answers with (e.g. 503).
  int liveOfStatus = 200;

  /// When set, `/livestreams?userId=` waits for it before answering.
  Completer<void>? holdLiveOf;

  /// Chat history answered by `GET /messages` (backend shape, `_id`).
  List<Map<String, Object?>> chatHistory = [];

  /// Commands answered by `GET /chat-commands`.
  List<Map<String, Object?>> chatCommands = [];

  /// Status `GET /messages` answers with (e.g. 503).
  int chatHistoryStatus = 200;

  /// When set, `GET /messages` waits for it before answering.
  Completer<void>? holdChatHistory;

  /// When set, `POST /messages` waits for it before answering.
  Completer<void>? holdSend;

  /// Status `POST /messages` answers with (e.g. 503).
  int sendStatus = 200;

  /// Bodies of the `POST /messages` calls, in order.
  final List<Map<String, Object?>> sentMessages = [];

  ResponseBody _chatError(int status) => jsonBody({
    'requestId': 'req-err',
    'success': false,
    'code': 1,
    'key': 'res_err_internal_server',
    'message': 'Server down',
  }, status: status);

  Future<ResponseBody?> _handleChat(RequestOptions o) async {
    if (o.path == ApiEndpoints.chatCommands && o.method == 'GET') {
      return okEnvelope(chatCommands);
    }
    if (o.path != ApiEndpoints.messages) return null;
    if (o.method == 'GET') {
      await holdChatHistory?.future;
      return chatHistoryStatus == 200
          ? okEnvelope(chatHistory)
          : _chatError(chatHistoryStatus);
    }
    if (o.method != 'POST') return null;
    final body = o.data as Map<String, Object?>;
    sentMessages.add(body);
    await holdSend?.future;
    if (sendStatus != 200) return _chatError(sendStatus);
    return okEnvelope({
      '_id': 'sent-${sentMessages.length}',
      'roomId': body['roomId'],
      'username': 'abcdef',
      'userId': 'u-1',
      'text': body['text'],
      'profilePicture': null,
      'timestamp': DateTime.now().toUtc().toIso8601String(),
      '__v': 0,
    });
  }

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    final chat = await _handleChat(o);
    if (chat != null) return chat;
    if (o.method == 'GET' && o.path == ApiEndpoints.livestreams) {
      await holdLiveOf?.future;
      if (liveOfStatus != 200) {
        return jsonBody({
          'requestId': 'req-err',
          'success': false,
          'code': 1,
          'key': 'res_err_internal_server',
          'message': 'Server down',
        }, status: liveOfStatus);
      }
      return okEnvelope(liveByUser[o.queryParameters['userId']]);
    }
    final isProfile =
        o.method == 'GET' &&
        o.path.startsWith('${ApiEndpoints.user}/') &&
        o.path != ApiEndpoints.userMe &&
        !o.path.startsWith('${ApiEndpoints.userMe}/');
    if (isProfile &&
        !users.containsKey(o.path.substring(ApiEndpoints.user.length + 1))) {
      return jsonBody({
        'requestId': 'req-404',
        'success': false,
        'code': 40401,
        'key': 'res_err_user_not_found',
        'message': 'User not found.',
      }, status: 404);
    }
    return super.handle(o);
  }
}

/// The signed-in app on Home with [players] standing in for the native
/// player.
Future<TestApp> pumpWatchHome(
  WidgetTester tester,
  FakeWatchServer server,
  FakePlayers players, {
  FakeScreenWake? wake,
  double width = 800,
  double height = 6000,
  double textScale = 1,
  List<Override> overrides = const [],
  AppConfig? config,
}) => pumpHome(
  tester,
  server,
  config: config,
  width: width,
  height: height,
  textScale: textScale,
  overrides: [
    playerFactoryProvider.overrideWithValue(players.create),
    // Never the real plugin in widget tests.
    screenWakeProvider.overrideWithValue((wake ?? FakeScreenWake()).call),
    ...overrides,
  ],
);

/// Pumps the signed-in app and opens the channel of [userId] the way a
/// card does (pushed above Home).
Future<TestApp> pumpChannel(
  WidgetTester tester,
  FakeWatchServer server,
  FakePlayers players,
  String userId, {
  FakeScreenWake? wake,
  double width = 800,
  double height = 6000,
  double textScale = 1,
  bool settle = true,
  List<Override> overrides = const [],
}) async {
  final app = await pumpWatchHome(
    tester,
    server,
    players,
    wake: wake,
    width: width,
    height: height,
    textScale: textScale,
    overrides: overrides,
  );
  unawaited(
    app.container.read(appRouterProvider).push(AppRoutes.channel(userId)),
  );
  if (settle) {
    await tester.pumpAndSettle();
  } else {
    await tester.pump();
  }
  return app;
}
