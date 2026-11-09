/// Realtime DM pushes, typed after `DmWsServerEvent` in `web/types/dm.ts`.
/// The gateway delivers the event type in the frame and the rest as `data`;
/// DM events travel on the user's own topic, which the server joins for the
/// socket, so (like web) nothing is subscribed here.
library;

import 'package:letslive_mobile/features/messages/data/dm_models.dart';

/// Event types web listens to (`DM_EVENT_TYPES` in `use-dm-realtime.ts`),
/// except `dm:read_receipt`, which web ignores too.
abstract final class DmEventType {
  static const String newMessage = 'dm:new_message';
  static const String messageEdited = 'dm:message_edited';
  static const String messageDeleted = 'dm:message_deleted';
  static const String userTyping = 'dm:user_typing';
  static const String userStoppedTyping = 'dm:user_stopped_typing';
  static const String userOnline = 'dm:user_online';
  static const String userOffline = 'dm:user_offline';
  static const String conversationUpdated = 'dm:conversation_updated';

  /// Every type the app subscribes to.
  static const List<String> all = [
    newMessage,
    messageEdited,
    messageDeleted,
    userTyping,
    userStoppedTyping,
    userOnline,
    userOffline,
    conversationUpdated,
  ];
}

/// Something that happened to the user's conversations.
sealed class DmEvent {
  const DmEvent();
}

/// A message arrived (also the user's own, echoed to their other devices).
class DmNewMessage extends DmEvent {
  const DmNewMessage(this.conversationId, this.message);

  final String conversationId;
  final DmMessage message;
}

/// A message's text changed (`dm:message_edited`).
class DmMessageEdited extends DmEvent {
  const DmMessageEdited({
    required this.conversationId,
    required this.messageId,
    required this.newText,
    required this.updatedAt,
  });

  final String conversationId;
  final String messageId;
  final String newText;
  final DateTime updatedAt;
}

/// A message was deleted (`dm:message_deleted`).
class DmMessageDeleted extends DmEvent {
  const DmMessageDeleted(this.conversationId, this.messageId);

  final String conversationId;
  final String messageId;
}

/// [username] started ([isTyping]) or stopped typing in a conversation.
class DmTyping extends DmEvent {
  const DmTyping({
    required this.conversationId,
    required this.username,
    required this.isTyping,
  });

  final String conversationId;
  final String username;
  final bool isTyping;
}

/// A user came online or went offline (`dm:user_online` / `dm:user_offline`).
class DmPresence extends DmEvent {
  const DmPresence({required this.userId, required this.isOnline});

  final String userId;
  final bool isOnline;
}

/// A conversation's fields changed (`dm:conversation_updated`), e.g. a rename.
class DmConversationUpdated extends DmEvent {
  const DmConversationUpdated(this.conversationId, this.update);

  final String conversationId;

  /// The changed fields, as in `Conversation.withUpdate`.
  final Map<String, Object?> update;
}

/// The socket reopened: pushes sent meanwhile were lost, refetch.
class DmReconnected extends DmEvent {
  const DmReconnected();
}

/// Parses a push of [type]. Returns null for a malformed payload (web
/// ignores those too); never throws.
DmEvent? parseDmEvent(String type, Object? data) {
  if (data is! Map<String, Object?>) return null;
  final conversationId = data['conversationId'];
  try {
    switch (type) {
      case DmEventType.newMessage:
        final message = data['message'];
        if (conversationId is! String || message is! Map<String, Object?>) {
          return null;
        }
        return DmNewMessage(conversationId, DmMessage.fromJson(message));
      case DmEventType.messageEdited:
        final messageId = data['messageId'];
        final newText = data['newText'];
        final updatedAt = data['updatedAt'];
        if (conversationId is! String ||
            messageId is! String ||
            newText is! String ||
            updatedAt is! String) {
          return null;
        }
        return DmMessageEdited(
          conversationId: conversationId,
          messageId: messageId,
          newText: newText,
          updatedAt: DateTime.parse(updatedAt),
        );
      case DmEventType.messageDeleted:
        final messageId = data['messageId'];
        if (conversationId is! String || messageId is! String) return null;
        return DmMessageDeleted(conversationId, messageId);
      case DmEventType.userTyping || DmEventType.userStoppedTyping:
        final username = data['username'];
        if (conversationId is! String || username is! String) return null;
        return DmTyping(
          conversationId: conversationId,
          username: username,
          isTyping: type == DmEventType.userTyping,
        );
      case DmEventType.userOnline || DmEventType.userOffline:
        final userId = data['userId'];
        if (userId is! String) return null;
        return DmPresence(
          userId: userId,
          isOnline: type == DmEventType.userOnline,
        );
      case DmEventType.conversationUpdated:
        final update = data['update'];
        if (conversationId is! String || update is! Map<String, Object?>) {
          return null;
        }
        return DmConversationUpdated(conversationId, update);
    }
  } on FormatException {
    return null;
  }
  return null;
}
