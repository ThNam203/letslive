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
/// Lines are appended in the order they arrive, so a join note or a notice
/// stamped with the device clock never ends up below a newer message. Only
/// with [insertByTime] (messages fetched after a reconnect, which can be older
/// than what is shown) is each new message placed before the first shown
/// message that is newer than it; other lines never move. The result keeps
/// the last [chatLineLimit] lines.
List<ChatLine> mergeChatLines(
  List<ChatLine> current,
  Iterable<ChatLine> incoming, {
  bool insertByTime = false,
}) {
  final seen = {
    for (final line in current)
      if (line is ChatMessageLine) line.message.id,
  };
  final merged = [...current];
  var changed = false;
  for (final line in incoming) {
    if (line is ChatMessageLine && !seen.add(line.message.id)) continue;
    changed = true;
    final at = insertByTime && line is ChatMessageLine
        ? merged.indexWhere(
            (l) => l is ChatMessageLine && l.at.isAfter(line.at),
          )
        : -1;
    if (at == -1) {
      merged.add(line);
    } else {
      merged.insert(at, line);
    }
  }
  if (!changed) return current;
  return merged.length > chatLineLimit
      ? merged.sublist(merged.length - chatLineLimit)
      : merged;
}
