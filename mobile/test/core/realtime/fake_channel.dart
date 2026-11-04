import 'dart:async';
import 'dart:convert';

import 'package:letslive_mobile/core/realtime/realtime_channel.dart';

/// Scriptable socket: the test decides when the handshake completes, what the
/// server sends and how the socket closes.
class FakeChannel implements RealtimeChannel {
  FakeChannel(this.uri, this.headers);

  final Uri uri;
  final Map<String, String> headers;
  final Completer<void> _ready = Completer<void>();
  final StreamController<String> _incoming = StreamController<String>();
  final List<Map<String, Object?>> sent = [];
  bool closedByClient = false;
  int? _closeCode;

  @override
  Future<void> get ready => _ready.future;

  @override
  Stream<String> get incoming => _incoming.stream;

  @override
  int? get closeCode => _closeCode;

  @override
  void send(String frame) {
    sent.add(jsonDecode(frame) as Map<String, Object?>);
  }

  @override
  Future<void> close() async {
    closedByClient = true;
    if (!_incoming.isClosed) await _incoming.close();
  }

  List<String> opsFor(String op) => [
    for (final f in sent)
      if (f['op'] == op) f['topic'] as String? ?? '',
  ];

  void open() => _ready.complete();

  void failHandshake(int? status) {
    // The manager must not rely on the stream when the handshake fails, but a
    // real socket also errors/closes it.
    _ready.completeError(RealtimeHandshakeException(status));
    _incoming.close();
  }

  void serverSends(Map<String, Object?> frame) =>
      _incoming.add(jsonEncode(frame));

  void serverSendsRaw(String raw) => _incoming.add(raw);

  void serverCloses([int? code]) {
    _closeCode = code;
    _incoming.close();
  }
}
