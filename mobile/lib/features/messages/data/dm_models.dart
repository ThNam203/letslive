/// Direct-message models, typed after `web/types/dm.ts` and the chat
/// service's `dto.ConversationResponse` / `dto.DmMessageResponse`. Only the
/// fields the app uses are modelled; those are parsed strictly (a mistyped
/// field throws [FormatException], which `ApiClient` reports as a malformed
/// response).
library;

import 'package:letslive_mobile/core/network/json_fields.dart';

/// A one-to-one chat or a named group.
enum ConversationType {
  dm,
  group;

  static ConversationType parse(String value) => switch (value) {
    'dm' => dm,
    'group' => group,
    _ => throw FormatException('Conversation.type: unknown "$value"'),
  };
}

/// What a message carries.
enum DmMessageType {
  text,
  image,

  /// A notice the server wrote (e.g. a member joined); shown centered.
  system;

  static DmMessageType parse(String value) => switch (value) {
    'text' => text,
    'image' => image,
    'system' => system,
    _ => throw FormatException('DmMessage.type: unknown "$value"'),
  };
}

/// A member of a conversation.
class Participant {
  const Participant({
    required this.userId,
    required this.username,
    this.profilePicture,
  });

  factory Participant.fromJson(Map<String, Object?> json) {
    final f = JsonFields('Participant', json);
    return Participant(
      userId: f.string('userId'),
      username: f.string('username'),
      profilePicture: f.optString('profilePicture'),
    );
  }

  final String userId;
  final String username;

  /// Absolute URL of the avatar; null when none.
  final String? profilePicture;
}

/// The preview of the newest message of a conversation.
class LastMessage {
  const LastMessage({
    required this.id,
    required this.senderId,
    required this.senderUsername,
    required this.text,
    required this.createdAt,
  });

  factory LastMessage.fromJson(Map<String, Object?> json) {
    final f = JsonFields('LastMessage', json);
    return LastMessage(
      id: f.string('_id'),
      senderId: f.string('senderId'),
      senderUsername: f.string('senderUsername'),
      text: f.string('text'),
      createdAt: f.date('createdAt'),
    );
  }

  final String id;
  final String senderId;
  final String senderUsername;
  final String text;
  final DateTime createdAt;
}

/// One conversation of the signed-in user.
class Conversation {
  const Conversation({
    required this.id,
    required this.type,
    required this.participants,
    required this.updatedAt,
    this.name,
    this.avatarUrl,
    this.lastMessage,
  });

  factory Conversation.fromJson(Map<String, Object?> json) {
    final f = JsonFields('Conversation', json);
    final last = json['lastMessage'];
    return Conversation(
      id: f.string('_id'),
      type: ConversationType.parse(f.string('type')),
      name: f.optString('name'),
      avatarUrl: f.optString('avatarUrl'),
      participants: parseJsonList(
        json['participants'],
        'Participant',
        Participant.fromJson,
      ),
      lastMessage: last == null
          ? null
          : parseJsonObject(last, 'LastMessage', LastMessage.fromJson),
      updatedAt: f.date('updatedAt'),
    );
  }

  /// `ParseData<Conversation>` for `ApiClient`.
  static Conversation parse(Object? data) =>
      parseJsonObject(data, 'Conversation', Conversation.fromJson);

  /// `ParseData<List<Conversation>>` for `ApiClient`.
  static List<Conversation> parseList(Object? data) =>
      parseJsonList(data, 'Conversation', Conversation.fromJson);

  final String id;
  final ConversationType type;

  /// Group name; null for direct messages and unnamed groups.
  final String? name;
  final String? avatarUrl;
  final List<Participant> participants;
  final LastMessage? lastMessage;
  final DateTime updatedAt;

  /// This conversation with [lastMessage] as its newest message.
  Conversation withLastMessage(LastMessage message) => Conversation(
    id: id,
    type: type,
    name: name,
    avatarUrl: avatarUrl,
    participants: participants,
    lastMessage: message,
    updatedAt: message.createdAt,
  );

  /// Applies a `dm:conversation_updated` payload (web `Partial<Conversation>`):
  /// only the keys present change. A key with a wrong type throws
  /// [FormatException].
  Conversation withUpdate(Map<String, Object?> update) {
    final f = JsonFields('Conversation.update', update);
    final last = update['lastMessage'];
    return Conversation(
      id: id,
      type: type,
      name: update.containsKey('name') ? f.optString('name') : name,
      avatarUrl: update.containsKey('avatarUrl')
          ? f.optString('avatarUrl')
          : avatarUrl,
      participants: update.containsKey('participants')
          ? parseJsonList(
              update['participants'],
              'Participant',
              Participant.fromJson,
            )
          : participants,
      lastMessage: update.containsKey('lastMessage')
          ? (last == null
                ? null
                : parseJsonObject(last, 'LastMessage', LastMessage.fromJson))
          : lastMessage,
      updatedAt: update.containsKey('updatedAt')
          ? f.date('updatedAt')
          : updatedAt,
    );
  }
}

/// One message of a conversation.
class DmMessage {
  const DmMessage({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.senderUsername,
    required this.type,
    required this.text,
    required this.imageUrls,
    required this.isDeleted,
    required this.createdAt,
    required this.updatedAt,
  });

  factory DmMessage.fromJson(Map<String, Object?> json) {
    final f = JsonFields('DmMessage', json);
    final urls = json['imageUrls'];
    if (urls is! List<Object?> || urls.any((u) => u is! String)) {
      throw FormatException('DmMessage.imageUrls: expected strings, got $urls');
    }
    return DmMessage(
      id: f.string('_id'),
      conversationId: f.string('conversationId'),
      senderId: f.string('senderId'),
      senderUsername: f.string('senderUsername'),
      type: DmMessageType.parse(f.string('type')),
      text: f.string('text'),
      imageUrls: List.unmodifiable(urls.cast<String>()),
      isDeleted: f.boolean('isDeleted'),
      createdAt: f.date('createdAt'),
      updatedAt: f.date('updatedAt'),
    );
  }

  /// `ParseData<DmMessage>` for `ApiClient`.
  static DmMessage parse(Object? data) =>
      parseJsonObject(data, 'DmMessage', DmMessage.fromJson);

  /// `ParseData<List<DmMessage>>` for `ApiClient`.
  static List<DmMessage> parseList(Object? data) =>
      parseJsonList(data, 'DmMessage', DmMessage.fromJson);

  final String id;
  final String conversationId;
  final String senderId;
  final String senderUsername;
  final DmMessageType type;

  /// Empty once [isDeleted].
  final String text;
  final List<String> imageUrls;
  final bool isDeleted;
  final DateTime createdAt;
  final DateTime updatedAt;

  /// Web shows "(edited)" when the two timestamps differ.
  bool get isEdited => updatedAt != createdAt;

  /// This message with new [text], edited at [at].
  DmMessage withText(String newText, DateTime at) =>
      _copy(text: newText, updatedAt: at);

  /// This message as the server keeps it after a delete.
  DmMessage asDeleted() => _copy(text: '', isDeleted: true);

  DmMessage _copy({String? text, bool? isDeleted, DateTime? updatedAt}) =>
      DmMessage(
        id: id,
        conversationId: conversationId,
        senderId: senderId,
        senderUsername: senderUsername,
        type: type,
        text: text ?? this.text,
        imageUrls: imageUrls,
        isDeleted: isDeleted ?? this.isDeleted,
        createdAt: createdAt,
        updatedAt: updatedAt ?? this.updatedAt,
      );
}
