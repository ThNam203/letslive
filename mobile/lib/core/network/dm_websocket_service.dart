import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:web_socket_channel/io.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import '../../models/conversation.dart';
import '../config/app_config.dart';

// ---------------------------------------------------------------------------
// DM WebSocket event types (must match backend / web client)
// ---------------------------------------------------------------------------

abstract final class DmClientEventType {
  static const sendMessage = 'dm:send_message';
  static const typingStart = 'dm:typing_start';
  static const typingStop = 'dm:typing_stop';
  static const markRead = 'dm:mark_read';
}

abstract final class DmServerEventType {
  static const newMessage = 'dm:new_message';
  static const messageEdited = 'dm:message_edited';
  static const messageDeleted = 'dm:message_deleted';
  static const userTyping = 'dm:user_typing';
  static const userStoppedTyping = 'dm:user_stopped_typing';
  static const readReceipt = 'dm:read_receipt';
  static const userOnline = 'dm:user_online';
  static const userOffline = 'dm:user_offline';
  static const conversationUpdated = 'dm:conversation_updated';
  static const sendFailed = 'dm:send_failed';
}

// ---------------------------------------------------------------------------
// Typed server events
// ---------------------------------------------------------------------------

sealed class DmServerEvent {
  final String type;
  const DmServerEvent(this.type);

  factory DmServerEvent.fromJson(Map<String, dynamic> json) {
    final type = json['type'] as String;
    return switch (type) {
      DmServerEventType.newMessage => DmNewMessageEvent.fromJson(json),
      DmServerEventType.messageEdited => DmMessageEditedEvent.fromJson(json),
      DmServerEventType.messageDeleted => DmMessageDeletedEvent.fromJson(json),
      DmServerEventType.userTyping ||
      DmServerEventType.userStoppedTyping => DmUserTypingEvent.fromJson(json),
      DmServerEventType.userOnline ||
      DmServerEventType.userOffline => DmPresenceEvent.fromJson(json),
      DmServerEventType.conversationUpdated =>
        DmConversationUpdatedEvent.fromJson(json),
      DmServerEventType.sendFailed => DmSendFailedEvent.fromJson(json),
      _ => DmUnknownEvent(type),
    };
  }
}

class DmNewMessageEvent extends DmServerEvent {
  final String conversationId;
  final DmMessage message;
  DmNewMessageEvent({required this.conversationId, required this.message})
    : super(DmServerEventType.newMessage);
  factory DmNewMessageEvent.fromJson(Map<String, dynamic> json) {
    return DmNewMessageEvent(
      conversationId: json['conversationId'] as String,
      message: DmMessage.fromJson(json['message'] as Map<String, dynamic>),
    );
  }
}

class DmMessageEditedEvent extends DmServerEvent {
  final String conversationId;
  final String messageId;
  final String newText;
  final String updatedAt;
  DmMessageEditedEvent({
    required this.conversationId,
    required this.messageId,
    required this.newText,
    required this.updatedAt,
  }) : super(DmServerEventType.messageEdited);
  factory DmMessageEditedEvent.fromJson(Map<String, dynamic> json) {
    return DmMessageEditedEvent(
      conversationId: json['conversationId'] as String,
      messageId: json['messageId'] as String,
      newText: json['newText'] as String,
      updatedAt: json['updatedAt'] as String,
    );
  }
}

class DmMessageDeletedEvent extends DmServerEvent {
  final String conversationId;
  final String messageId;
  DmMessageDeletedEvent({required this.conversationId, required this.messageId})
    : super(DmServerEventType.messageDeleted);
  factory DmMessageDeletedEvent.fromJson(Map<String, dynamic> json) {
    return DmMessageDeletedEvent(
      conversationId: json['conversationId'] as String,
      messageId: json['messageId'] as String,
    );
  }
}

class DmUserTypingEvent extends DmServerEvent {
  final String conversationId;
  final String userId;
  final String username;
  DmUserTypingEvent({
    required String type,
    required this.conversationId,
    required this.userId,
    required this.username,
  }) : super(type);
  factory DmUserTypingEvent.fromJson(Map<String, dynamic> json) {
    return DmUserTypingEvent(
      type: json['type'] as String,
      conversationId: json['conversationId'] as String,
      userId: json['userId'] as String,
      username: json['username'] as String,
    );
  }
}

class DmPresenceEvent extends DmServerEvent {
  final String userId;
  DmPresenceEvent({required String type, required this.userId}) : super(type);
  factory DmPresenceEvent.fromJson(Map<String, dynamic> json) {
    return DmPresenceEvent(
      type: json['type'] as String,
      userId: json['userId'] as String,
    );
  }
}

class DmConversationUpdatedEvent extends DmServerEvent {
  final String conversationId;
  final Map<String, dynamic> update;
  DmConversationUpdatedEvent({
    required this.conversationId,
    required this.update,
  }) : super(DmServerEventType.conversationUpdated);
  factory DmConversationUpdatedEvent.fromJson(Map<String, dynamic> json) {
    return DmConversationUpdatedEvent(
      conversationId: json['conversationId'] as String,
      update: json['update'] as Map<String, dynamic>? ?? {},
    );
  }
}

class DmSendFailedEvent extends DmServerEvent {
  final String key;
  final String? message;
  DmSendFailedEvent({required this.key, this.message})
    : super(DmServerEventType.sendFailed);
  factory DmSendFailedEvent.fromJson(Map<String, dynamic> json) {
    return DmSendFailedEvent(
      key: json['key'] as String? ?? '',
      message: json['message'] as String?,
    );
  }
}

class DmUnknownEvent extends DmServerEvent {
  DmUnknownEvent(super.type);
}

// ---------------------------------------------------------------------------
// DM WebSocket Service (matches web's useDmWebSocket hook)
// ---------------------------------------------------------------------------

class DmWebSocketService {
  WebSocketChannel? _channel;
  final _eventController = StreamController<DmServerEvent>.broadcast();
  Timer? _reconnectTimer;
  bool _disposed = false;
  bool _isConnecting = false;

  final CookieJar _cookieJar;

  int _reconnectDelay = _initialReconnectDelay;
  static const _initialReconnectDelay = 1000; // ms
  static const _maxReconnectDelay = 30000; // ms

  // Typing indicator timeout management
  final _typingTimeouts = <String, Timer>{};
  static const _typingTimeoutMs = 5000;

  // Typing state for local user (debounce outgoing typing events)
  bool _isLocalTyping = false;
  Timer? _localTypingStopTimer;

  DmWebSocketService(this._cookieJar);

  /// Stream of parsed server events.
  Stream<DmServerEvent> get events => _eventController.stream;

  bool get isConnected => _channel != null && !_disposed;

  Future<void> connect() async {
    if (_disposed || _isConnecting) return;
    _isConnecting = true;

    try {
      final uri = Uri.parse(AppConfig.dmWsUrl);

      // Extract cookies from the shared cookie jar for WebSocket auth
      final cookies = await _cookieJar.loadForRequest(
        Uri.parse(AppConfig.apiUrl),
      );
      final cookieHeader = cookies.map((c) => '${c.name}=${c.value}').join('; ');

      _channel = IOWebSocketChannel.connect(
        uri,
        headers: cookieHeader.isNotEmpty ? {'Cookie': cookieHeader} : null,
      );

      _channel!.stream.listen(
        (data) {
          try {
            final json = jsonDecode(data as String) as Map<String, dynamic>;
            final event = DmServerEvent.fromJson(json);
            _handleTypingTimeouts(event);
            _eventController.add(event);
          } catch (_) {}
        },
        onError: (_) {
          _isConnecting = false;
          _scheduleReconnect();
        },
        onDone: () {
          _isConnecting = false;
          _channel = null;
          _scheduleReconnect();
        },
      );

      _reconnectDelay = _initialReconnectDelay;
      _isConnecting = false;
    } catch (_) {
      _isConnecting = false;
      _scheduleReconnect();
    }
  }

  /// Send a raw client event to the server.
  void send(Map<String, dynamic> event) {
    if (_channel == null) return;
    try {
      _channel!.sink.add(jsonEncode(event));
    } catch (_) {}
  }

  /// Notify the server that the local user started typing.
  /// Call this on every keystroke; debouncing is handled internally.
  void handleTyping({
    required String conversationId,
    required String username,
  }) {
    if (!_isLocalTyping) {
      _isLocalTyping = true;
      send({
        'type': DmClientEventType.typingStart,
        'conversationId': conversationId,
        'username': username,
      });
    }

    _localTypingStopTimer?.cancel();
    _localTypingStopTimer = Timer(const Duration(seconds: 2), () {
      _isLocalTyping = false;
      send({
        'type': DmClientEventType.typingStop,
        'conversationId': conversationId,
        'username': username,
      });
    });
  }

  /// Stop the local typing indicator immediately (e.g. on message send).
  void stopTyping({required String conversationId, required String username}) {
    if (_isLocalTyping) {
      _isLocalTyping = false;
      _localTypingStopTimer?.cancel();
      send({
        'type': DmClientEventType.typingStop,
        'conversationId': conversationId,
        'username': username,
      });
    }
  }

  // Auto-expire remote typing indicators after timeout
  void _handleTypingTimeouts(DmServerEvent event) {
    if (event is DmUserTypingEvent) {
      final key = '${event.conversationId}:${event.username}';
      if (event.type == DmServerEventType.userTyping) {
        _typingTimeouts[key]?.cancel();
        _typingTimeouts[key] = Timer(
          const Duration(milliseconds: _typingTimeoutMs),
          () {
            _typingTimeouts.remove(key);
            _eventController.add(
              DmUserTypingEvent(
                type: DmServerEventType.userStoppedTyping,
                conversationId: event.conversationId,
                userId: event.userId,
                username: event.username,
              ),
            );
          },
        );
      } else {
        // Stopped typing — cancel timeout
        _typingTimeouts[key]?.cancel();
        _typingTimeouts.remove(key);
      }
    }
  }

  void _scheduleReconnect() {
    if (_disposed) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(Duration(milliseconds: _reconnectDelay), () {
      _reconnectDelay = min(_reconnectDelay * 2, _maxReconnectDelay);
      if (!_disposed) connect();
    });
  }

  void disconnect() {
    _localTypingStopTimer?.cancel();
    _reconnectTimer?.cancel();
    for (final t in _typingTimeouts.values) {
      t.cancel();
    }
    _typingTimeouts.clear();
    _channel?.sink.close();
    _channel = null;
  }

  void dispose() {
    _disposed = true;
    disconnect();
    _eventController.close();
  }
}
