import 'package:letslive_mobile/features/watch/chat/data/chat_message.dart';

/// Most lines the panel keeps; older ones fall off the top.
const int chatLineLimit = 200;

/// One row of the chat panel.
sealed class ChatLine {
  const ChatLine(this.at);

  /// When it happened; lines are shown in this order.
  final DateTime at;
}

/// Something a viewer said.
class ChatMessageLine extends ChatLine {
  ChatMessageLine(this.message) : super(message.sentAt);

  final ChatMessage message;
}

/// A viewer joined or left the room (`member.joined` / `member.left`, which
/// carry only the user id).
class ChatMemberLine extends ChatLine {
  const ChatMemberLine({
    required this.userId,
    required this.joined,
    required DateTime at,
  }) : super(at);

  final String userId;
  final bool joined;
}

/// A note only the sender sees: `/help` output or a command error.
class ChatNoticeLine extends ChatLine {
  const ChatNoticeLine({required this.text, required DateTime at}) : super(at);

  final String text;
}

/// [current] plus the [incoming] lines, with the rules that keep the same
/// message from showing twice: a message whose id is already there is
/// dropped, whether it came from history, a push, a push repeated after a
/// reconnect, or the send response racing its own push.
///
/// Lines normally arrive in order and are appended. Messages fetched after a
/// reconnect can be older than lines already shown, so then the whole list is
/// put back in time order (equal times keep their order). The result keeps
/// the last [chatLineLimit] lines.
List<ChatLine> mergeChatLines(
  List<ChatLine> current,
  Iterable<ChatLine> incoming,
) {
  final seen = {
    for (final line in current)
      if (line is ChatMessageLine) line.message.id,
  };
  final added = <ChatLine>[];
  for (final line in incoming) {
    if (line is ChatMessageLine && !seen.add(line.message.id)) continue;
    added.add(line);
  }
  if (added.isEmpty) return current;

  var merged = [...current, ...added];
  final last = current.isEmpty ? null : current.last.at;
  if (last != null && added.any((line) => line.at.isBefore(last))) {
    final order = {for (var i = 0; i < merged.length; i++) merged[i]: i};
    merged = [...merged]
      ..sort((a, b) {
        final byTime = a.at.compareTo(b.at);
        return byTime != 0 ? byTime : order[a]!.compareTo(order[b]!);
      });
  }
  return merged.length > chatLineLimit
      ? merged.sublist(merged.length - chatLineLimit)
      : merged;
}
