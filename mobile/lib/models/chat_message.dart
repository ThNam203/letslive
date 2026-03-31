class ChatMessage {
  final String? id;
  final String type; // "message", "join", "leave"
  final String userId;
  final String username;
  final String text;
  final int? timestamp;

  const ChatMessage({
    this.id,
    required this.type,
    required this.userId,
    required this.username,
    required this.text,
    this.timestamp,
  });

  bool get isJoin => type == 'join';
  bool get isLeave => type == 'leave';
  bool get isMessage => type == 'message';

  Map<String, dynamic> toSendJson(String roomId) {
    return {
      'type': type,
      'roomId': roomId,
      'userId': userId,
      'username': username,
      'text': text,
    };
  }

  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id'] as String? ?? json['_id'] as String?,
      type: json['type'] as String? ?? 'message',
      userId: json['userId'] as String,
      username: json['username'] as String,
      text: json['text'] as String? ?? '',
      timestamp: json['timestamp'] as int?,
    );
  }
}
