import 'dart:async';

import 'package:web_socket_channel/io.dart';

/// The handshake failed; [statusCode] is the HTTP status when the server
/// answered the upgrade with one (e.g. 401), null for transport failures.
class RealtimeHandshakeException implements Exception {
  const RealtimeHandshakeException(this.statusCode);

  final int? statusCode;

  @override
  String toString() => 'RealtimeHandshakeException($statusCode)';
}

/// Minimal socket surface the manager needs; tests inject a fake.
abstract interface class RealtimeChannel {
  /// Completes when the upgrade succeeded, or errors with
  /// [RealtimeHandshakeException].
  Future<void> get ready;

  /// Text frames. Completes (or errors) when the socket closes.
  Stream<String> get incoming;

  /// Close code once closed, if the peer sent one.
  int? get closeCode;

  void send(String frame);

  Future<void> close();
}

typedef RealtimeChannelFactory = RealtimeChannel Function(
  Uri uri,
  Map<String, String> headers,
);

/// Production factory over `dart:io`; [headers] carry the auth cookies.
RealtimeChannel connectIoChannel(Uri uri, Map<String, String> headers) =>
    _IoRealtimeChannel(IOWebSocketChannel.connect(uri, headers: headers));

class _IoRealtimeChannel implements RealtimeChannel {
  _IoRealtimeChannel(this._channel);

  final IOWebSocketChannel _channel;
  static final RegExp _status = RegExp(r'HTTP status code: (\d{3})');

  @override
  Future<void> get ready async {
    try {
      await _channel.ready;
    } on Object catch (e) {
      // dart:io reports "...was not upgraded to websocket, HTTP status code:
      // 401" inside web_socket_channel's own exception type.
      final match = _status.firstMatch(e.toString());
      throw RealtimeHandshakeException(
        match == null ? null : int.parse(match.group(1)!),
      );
    }
  }

  @override
  Stream<String> get incoming =>
      _channel.stream.where((m) => m is String).cast<String>();

  @override
  int? get closeCode => _channel.closeCode;

  @override
  void send(String frame) => _channel.sink.add(frame);

  @override
  Future<void> close() => _channel.sink.close();
}
