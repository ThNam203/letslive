import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../../models/chat_message.dart';
import '../config/app_config.dart';

class LiveChatService {
  WebSocketChannel? _channel;
  final _messageController = StreamController<ChatMessage>.broadcast();
  Timer? _reconnectTimer;
  bool _disposed = false;

  String? _roomId;
  String? _userId;
  String? _username;

  Stream<ChatMessage> get messages => _messageController.stream;

  void connect({
    required String roomId,
    required String userId,
    required String username,
  }) {
    _roomId = roomId;
    _userId = userId;
    _username = username;
    _connectInternal();
  }

  void _connectInternal() {
    if (_disposed) return;

    try {
      final uri = Uri.parse(AppConfig.wsUrl);
      _channel = WebSocketChannel.connect(uri);

      _channel!.stream.listen(
        (data) {
          try {
            final json = jsonDecode(data as String) as Map<String, dynamic>;
            final message = ChatMessage.fromJson(json);
            _messageController.add(message);
          } catch (_) {}
        },
        onError: (_) => _scheduleReconnect(),
        onDone: () => _scheduleReconnect(),
      );

      // Send join message
      _send(
        ChatMessage(
          type: 'join',
          userId: _userId!,
          username: _username!,
          text: '',
        ),
      );
    } catch (_) {
      _scheduleReconnect();
    }
  }

  void sendMessage(String text) {
    if (_roomId == null || _userId == null || _username == null) return;

    _send(
      ChatMessage(
        type: 'message',
        userId: _userId!,
        username: _username!,
        text: text,
      ),
    );
  }

  void _send(ChatMessage message) {
    if (_channel == null || _roomId == null) return;

    try {
      _channel!.sink.add(jsonEncode(message.toSendJson(_roomId!)));
    } catch (_) {}
  }

  void _scheduleReconnect() {
    if (_disposed) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(const Duration(seconds: 3), () {
      if (!_disposed) _connectInternal();
    });
  }

  void disconnect() {
    if (_roomId != null && _userId != null && _username != null) {
      _send(
        ChatMessage(
          type: 'leave',
          userId: _userId!,
          username: _username!,
          text: '',
        ),
      );
    }

    _reconnectTimer?.cancel();
    _channel?.sink.close();
    _channel = null;
  }

  void dispose() {
    _disposed = true;
    disconnect();
    _messageController.close();
  }
}
