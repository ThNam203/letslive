/// Riverpod providers of the live chat panel on the channel screen.
///
/// `ChatController` keeps one room's lines: the history from `GET /messages`
/// merged with the realtime events of topic `room:<streamerId>`
/// (`chat.message`, `member.joined`, `member.left`; payloads in
/// `backend/chat/events` and `backend/realtime`). Every line is merged by
/// message id ([mergeChatLines]), so a message that arrives through the
/// history, a push, a push repeated after a reconnect and the send response
/// shows once. After a reconnect (background, lost network) the history is
/// fetched again: pushes sent while the socket was down only exist there.
/// Web does the same, but throws its live lines away instead of merging.
///
/// Both providers are `autoDispose`: leaving the screen unsubscribes the
/// topic and drops the handlers, and logout drops the lines.
library;

import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/chat_parser/chat_command.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/realtime/realtime_frame.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';
import 'package:letslive_mobile/features/watch/chat/data/chat_repository.dart';
import 'package:letslive_mobile/features/watch/chat/domain/chat_line.dart';

/// Livestream chat REST calls.
final chatRepositoryProvider = Provider<ChatRepository>(
  (ref) => ChatRepository(ref.watch(apiClientProvider)),
);

/// The lines of one room (the streamer's user id).
final chatControllerProvider = AsyncNotifierProvider.autoDispose
    .family<ChatController, List<ChatLine>, String>(
      ChatController.new,
      retry: noRetry,
    );

/// The room's channel commands plus the viewer's own. A failure is an error
/// value; the input then offers only the built-ins.
final chatCommandsProvider = FutureProvider.autoDispose
    .family<List<ChatCommand>, String>(
      (ref, roomId) async => unwrapList(
        await ref.watch(chatRepositoryProvider).commandsFor(roomId),
      ),
      retry: noRetry,
    );

/// State and actions of one room's chat (see the library doc).
class ChatController extends AsyncNotifier<List<ChatLine>> {
  ChatController(this.roomId);

  /// The room, which is the streamer's user id.
  final String roomId;

  // Events that arrive before the history has loaded; merged into it.
  final List<ChatLine> _early = [];

  @override
  Future<List<ChatLine>> build() async {
    _early.clear();
    final realtime = ref.read(realtimeManagerProvider);
    final topic = realtimeRoomTopic(roomId);
    // Subscribe before fetching so nothing said in between is lost.
    final cleanups = [
      realtime.subscribe(topic),
      realtime.onEvent(RealtimeEvent.chatMessage, _onChatMessage, topic: topic),
      realtime.onEvent(
        RealtimeEvent.memberJoined,
        (data) => _onMember(data, joined: true),
        topic: topic,
      ),
      realtime.onEvent(
        RealtimeEvent.memberLeft,
        (data) => _onMember(data, joined: false),
        topic: topic,
      ),
      realtime.onReconnect(() => unawaited(_catchUp())),
    ];
    ref.onDispose(() {
      for (final cleanup in cleanups) {
        cleanup();
      }
    });

    final history = unwrapList(
      await ref.read(chatRepositoryProvider).history(roomId),
    );
    final lines = mergeChatLines(const [], [
      for (final m in history) ChatMessageLine(m),
    ]);
    final merged = mergeChatLines(lines, _early);
    _early.clear();
    return merged;
  }

  /// Says [text] (already expanded from any command) in the room. On success
  /// the line is added at once (merged by id with its own push); a failure is
  /// returned for the caller to show, nothing is added and the caller keeps
  /// the text.
  Future<ApiResponse<ChatMessage>> send(String text) async {
    final response = await ref.read(chatRepositoryProvider).send(roomId, text);
    final sent = response.data;
    if (response.success && sent != null && ref.mounted) {
      _add(ChatMessageLine(sent));
    }
    return response;
  }

  /// Adds a line only the sender sees (`/help` text, a command error).
  void addNotice(String text) =>
      _add(ChatNoticeLine(text: text, at: DateTime.now()));

  void _onChatMessage(Object? data) {
    if (data is! Map<String, Object?>) return;
    try {
      _add(ChatMessageLine(ChatMessage.fromJson(data)));
    } on FormatException {
      // Not a chat line this app understands; web ignores it too.
    }
  }

  void _onMember(Object? data, {required bool joined}) {
    final userId = data is Map<String, Object?> ? data['userId'] : null;
    if (userId is! String) return;
    _add(ChatMemberLine(userId: userId, joined: joined, at: DateTime.now()));
  }

  /// After a reconnect: pull what was said while the socket was down. A
  /// failed fetch changes nothing; the next reconnect tries again.
  Future<void> _catchUp() async {
    final response = await ref.read(chatRepositoryProvider).history(roomId);
    final history = response.data;
    if (!ref.mounted || !response.success || history == null) return;
    _addAll([for (final m in history) ChatMessageLine(m)]);
  }

  void _add(ChatLine line) => _addAll([line]);

  void _addAll(List<ChatLine> lines) {
    final current = state.value;
    if (current == null) {
      _early.addAll(lines);
    } else {
      state = AsyncData(mergeChatLines(current, lines));
    }
  }
}
