import 'dart:async';

import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_repository.dart';

ApiResponse<T> okResponse<T>(T data) => ApiResponse<T>(
  success: true,
  statusCode: 200,
  code: 0,
  key: '',
  message: '',
  requestId: 'r',
  data: data,
);

ApiResponse<T> failedResponse<T>({int status = 503}) => ApiResponse<T>(
  success: false,
  statusCode: status,
  code: 1,
  key: 'res_err_internal_server',
  message: 'Server down',
  requestId: 'r',
);

/// A message as the history endpoint and the send response shape it.
Map<String, Object?> historyJson(
  String id,
  String text, {
  String userId = 'u-2',
  String username = 'viewer',
  String at = '2026-11-04T10:00:00Z',
}) => {
  '_id': id,
  'roomId': 'room-1',
  'username': username,
  'userId': userId,
  'text': text,
  'profilePicture': null,
  'timestamp': at,
  '__v': 0,
};

/// A `chat.message` push payload (`events.ChatMessage`).
Map<String, Object?> pushJson(
  String id,
  String text, {
  String userId = 'u-2',
  String username = 'viewer',
  int at = 1793786400000,
}) => {
  'id': id,
  'userId': userId,
  'username': username,
  'text': text,
  'profilePicture': null,
  'timestamp': at,
};

ChatMessage message(
  String id,
  String text, {
  String at = '2026-11-04T10:00:00Z',
}) => ChatMessage.fromJson(historyJson(id, text, at: at));

/// Scriptable [ChatRepository]: [histories] are answered in order (the last
/// one repeats), [sendResult] answers every send.
class FakeChatRepository implements ChatRepository {
  FakeChatRepository({List<ApiResponse<List<ChatMessage>>>? histories})
    : histories = histories ?? [okResponse(<ChatMessage>[])];

  final List<ApiResponse<List<ChatMessage>>> histories;
  ApiResponse<List<ChatCommand>> commands = okResponse(<ChatCommand>[]);
  ApiResponse<ChatMessage>? sendResult;

  /// When set, `history` waits for it.
  Completer<void>? holdHistory;

  int historyCalls = 0;
  final List<String> sent = [];

  @override
  Future<ApiResponse<List<ChatMessage>>> history(String roomId) async {
    final index = historyCalls < histories.length
        ? historyCalls
        : histories.length - 1;
    historyCalls++;
    await holdHistory?.future;
    return histories[index];
  }

  @override
  Future<ApiResponse<ChatMessage>> send(String roomId, String text) async {
    sent.add(text);
    return sendResult ?? failedResponse<ChatMessage>();
  }

  @override
  Future<ApiResponse<List<ChatCommand>>> commandsFor(String roomId) async =>
      commands;
}
