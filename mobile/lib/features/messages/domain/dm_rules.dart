/// Pure DM rules: who may change a message, how messages merge, and the
/// timings shared with web.
library;

import 'package:letslive_mobile/features/messages/data/dm_models.dart';

/// Conversations per page (web `CONVERSATIONS_PAGE_SIZE`).
const int dmConversationsPageSize = 20;

/// Messages per page (web `MESSAGES_PAGE_SIZE`, also the server default).
const int dmMessagesPageSize = 50;

/// How long a "typing" hint stays without a refresh
/// (`TYPING_INDICATOR_TIMEOUT_MS` in `web/hooks/use-dm-realtime.ts`).
const Duration dmTypingTimeout = Duration(seconds: 5);

/// How long the user must stop typing before the others are told so
/// (the 2000 ms timer in `web/.../message-input.tsx`).
const Duration dmTypingStopDelay = Duration(seconds: 2);

/// Whether [userId] may edit [message]: only the sender, only text that is
/// not deleted (the server refuses anything else).
bool canEditMessage(DmMessage message, String? userId) =>
    userId != null &&
    message.senderId == userId &&
    !message.isDeleted &&
    message.type == DmMessageType.text;

/// Whether [userId] may delete [message]: only the sender, not a system
/// notice, not twice.
bool canDeleteMessage(DmMessage message, String? userId) =>
    userId != null &&
    message.senderId == userId &&
    !message.isDeleted &&
    message.type != DmMessageType.system;

/// [current] plus [incoming], oldest first and without duplicates. A message
/// already known is replaced by the incoming copy, which is the fresher one
/// (it may carry an edit or a delete). The same message arrives from the
/// history, a push, a push repeated after a reconnect and the send response,
/// so merging by id is what keeps it from showing twice.
List<DmMessage> mergeDmMessages(
  List<DmMessage> current,
  Iterable<DmMessage> incoming,
) {
  final byId = {for (final m in current) m.id: m};
  for (final m in incoming) {
    byId[m.id] = m;
  }
  // Ids are time-ordered object ids, so they break ties of one millisecond.
  return byId.values.toList()..sort((a, b) {
    final byTime = a.createdAt.compareTo(b.createdAt);
    return byTime != 0 ? byTime : a.id.compareTo(b.id);
  });
}

/// The member of a direct message that is not [myId]; null in a group or
/// when the other member left.
Participant? otherParticipant(Conversation conversation, String? myId) {
  if (conversation.type != ConversationType.dm) return null;
  for (final p in conversation.participants) {
    if (p.userId != myId) return p;
  }
  return null;
}
