import 'package:letslive_mobile/core/network/json_fields.dart';

/// One chat line of a livestream room, typed after `ReceivedMessage` in
/// `web/types/message.ts`.
///
/// The same message arrives in three shapes that differ only in the id key
/// and the timestamp format: the history list and the send response
/// (`_id`, ISO string; `dto.LiveMessageResponse`) and the `chat.message` push
/// (`id`, epoch milliseconds; `events.ChatMessage`). The id is the object id
/// the server assigned, so it identifies a message across all three.
class ChatMessage {
  const ChatMessage({
    required this.id,
    required this.userId,
    required this.username,
    required this.text,
    required this.sentAt,
    this.profilePicture,
  });

  /// Reads any of the three shapes; throws [FormatException] otherwise.
  factory ChatMessage.fromJson(Map<String, Object?> json) {
    final f = JsonFields('ChatMessage', json);
    final stamp = json['timestamp'];
    final DateTime sentAt;
    if (stamp is int) {
      sentAt = DateTime.fromMillisecondsSinceEpoch(stamp);
    } else {
      sentAt = f.date('timestamp');
    }
    return ChatMessage(
      id: json.containsKey('id') ? f.string('id') : f.string('_id'),
      userId: f.string('userId'),
      username: f.string('username'),
      text: f.string('text'),
      profilePicture: f.optString('profilePicture'),
      sentAt: sentAt,
    );
  }

  /// `ParseData<ChatMessage>` for `ApiClient`.
  static ChatMessage parse(Object? data) =>
      parseJsonObject(data, 'ChatMessage', ChatMessage.fromJson);

  /// `ParseData<List<ChatMessage>>` for `ApiClient`.
  static List<ChatMessage> parseList(Object? data) =>
      parseJsonList(data, 'ChatMessage', ChatMessage.fromJson);

  final String id;
  final String userId;
  final String username;
  final String text;
  final DateTime sentAt;

  /// Absolute URL of the sender's avatar; null when none.
  final String? profilePicture;
}
