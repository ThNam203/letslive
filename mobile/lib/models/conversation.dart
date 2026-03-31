enum ConversationType {
  dm('dm'),
  group('group');

  final String value;
  const ConversationType(this.value);

  factory ConversationType.fromString(String value) {
    return ConversationType.values.firstWhere(
      (e) => e.value == value,
      orElse: () => ConversationType.dm,
    );
  }
}

enum DmMessageType {
  text('text'),
  image('image'),
  system('system');

  final String value;
  const DmMessageType(this.value);

  factory DmMessageType.fromString(String value) {
    return DmMessageType.values.firstWhere(
      (e) => e.value == value,
      orElse: () => DmMessageType.text,
    );
  }
}

enum ParticipantRole {
  owner('owner'),
  admin('admin'),
  member('member');

  final String value;
  const ParticipantRole(this.value);

  factory ParticipantRole.fromString(String value) {
    return ParticipantRole.values.firstWhere(
      (e) => e.value == value,
      orElse: () => ParticipantRole.member,
    );
  }
}

class ConversationParticipant {
  final String userId;
  final String username;
  final String? displayName;
  final String? profilePicture;
  final ParticipantRole role;
  final String joinedAt;
  final String? lastReadMessageId;
  final bool isMuted;

  const ConversationParticipant({
    required this.userId,
    required this.username,
    this.displayName,
    this.profilePicture,
    required this.role,
    required this.joinedAt,
    this.lastReadMessageId,
    this.isMuted = false,
  });

  factory ConversationParticipant.fromJson(Map<String, dynamic> json) {
    return ConversationParticipant(
      userId: json['userId'] as String,
      username: json['username'] as String,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
      role: ParticipantRole.fromString(json['role'] as String? ?? 'member'),
      joinedAt: json['joinedAt'] as String,
      lastReadMessageId: json['lastReadMessageId'] as String?,
      isMuted: json['isMuted'] as bool? ?? false,
    );
  }
}

class LastMessage {
  final String id;
  final String senderId;
  final String senderUsername;
  final String text;
  final String createdAt;

  const LastMessage({
    required this.id,
    required this.senderId,
    required this.senderUsername,
    required this.text,
    required this.createdAt,
  });

  factory LastMessage.fromJson(Map<String, dynamic> json) {
    return LastMessage(
      id: json['_id'] as String,
      senderId: json['senderId'] as String,
      senderUsername: json['senderUsername'] as String,
      text: json['text'] as String,
      createdAt: json['createdAt'] as String,
    );
  }
}

class Conversation {
  final String id;
  final ConversationType type;
  final String? name;
  final String? avatarUrl;
  final String createdBy;
  final List<ConversationParticipant> participants;
  final LastMessage? lastMessage;
  final String createdAt;
  final String updatedAt;

  const Conversation({
    required this.id,
    required this.type,
    this.name,
    this.avatarUrl,
    required this.createdBy,
    required this.participants,
    this.lastMessage,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Conversation.fromJson(Map<String, dynamic> json) {
    return Conversation(
      id: json['_id'] as String,
      type: ConversationType.fromString(json['type'] as String? ?? 'dm'),
      name: json['name'] as String?,
      avatarUrl: json['avatarUrl'] as String?,
      createdBy: json['createdBy'] as String,
      participants:
          (json['participants'] as List<dynamic>?)
              ?.map(
                (e) =>
                    ConversationParticipant.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          [],
      lastMessage: json['lastMessage'] != null
          ? LastMessage.fromJson(json['lastMessage'] as Map<String, dynamic>)
          : null,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
    );
  }
}

class ReadReceipt {
  final String userId;
  final String readAt;

  const ReadReceipt({required this.userId, required this.readAt});

  factory ReadReceipt.fromJson(Map<String, dynamic> json) {
    return ReadReceipt(
      userId: json['userId'] as String,
      readAt: json['readAt'] as String,
    );
  }
}

class DmMessage {
  final String id;
  final String conversationId;
  final String senderId;
  final String senderUsername;
  final DmMessageType type;
  final String text;
  final List<String>? imageUrls;
  final String? replyTo;
  final bool isDeleted;
  final List<ReadReceipt> readBy;
  final String createdAt;
  final String updatedAt;

  const DmMessage({
    required this.id,
    required this.conversationId,
    required this.senderId,
    required this.senderUsername,
    required this.type,
    required this.text,
    this.imageUrls,
    this.replyTo,
    this.isDeleted = false,
    required this.readBy,
    required this.createdAt,
    required this.updatedAt,
  });

  factory DmMessage.fromJson(Map<String, dynamic> json) {
    return DmMessage(
      id: json['_id'] as String,
      conversationId: json['conversationId'] as String,
      senderId: json['senderId'] as String,
      senderUsername: json['senderUsername'] as String,
      type: DmMessageType.fromString(json['type'] as String? ?? 'text'),
      text: json['text'] as String,
      imageUrls: (json['imageUrls'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      replyTo: json['replyTo'] as String?,
      isDeleted: json['isDeleted'] as bool? ?? false,
      readBy:
          (json['readBy'] as List<dynamic>?)
              ?.map((e) => ReadReceipt.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
    );
  }
}
