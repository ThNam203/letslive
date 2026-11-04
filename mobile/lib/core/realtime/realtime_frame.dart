import 'dart:convert';

/// Event types published by the realtime service (`web/constant/realtime.ts`).
abstract final class RealtimeEvent {
  static const String notificationCreated = 'notification.created';
  static const String chatMessage = 'chat.message';
  static const String memberJoined = 'member.joined';
  static const String memberLeft = 'member.left';
}

/// Topic of a livestream chat room (`realtimeRoomTopic` on the web).
String realtimeRoomTopic(String roomId) => 'room:$roomId';

/// Frame sent by the server (`backend/realtime/protocol/protocol.go`).
sealed class RealtimeFrame {
  const RealtimeFrame();
}

class RealtimeEventFrame extends RealtimeFrame {
  const RealtimeEventFrame({
    required this.topic,
    required this.type,
    required this.data,
  });

  final String topic;
  final String type;
  final Object? data;
}

class RealtimeErrorFrame extends RealtimeFrame {
  const RealtimeErrorFrame(this.code);

  /// One of invalid_topic, forbidden_topic, too_many_subscriptions,
  /// bad_frame, unavailable.
  final String code;
}

class RealtimePongFrame extends RealtimeFrame {
  const RealtimePongFrame();
}

/// Parses a server frame. Returns null for anything malformed or unknown;
/// never throws.
RealtimeFrame? parseRealtimeFrame(Object? raw) {
  if (raw is! String) return null;
  final Object? decoded;
  try {
    decoded = jsonDecode(raw);
  } on FormatException {
    return null;
  }
  if (decoded is! Map<String, Object?>) return null;
  switch (decoded['op']) {
    case 'pong':
      return const RealtimePongFrame();
    case 'error':
      final code = decoded['code'];
      return code is String ? RealtimeErrorFrame(code) : null;
    case 'event':
      final topic = decoded['topic'];
      final type = decoded['type'];
      if (topic is! String || type is! String) return null;
      return RealtimeEventFrame(
        topic: topic,
        type: type,
        data: decoded['data'],
      );
    default:
      return null;
  }
}

/// Encodes a client frame (`subscribe`, `unsubscribe` or `ping`).
String encodeRealtimeFrame(String op, {String? topic}) =>
    jsonEncode(<String, String>{'op': op, 'topic': ?topic});
