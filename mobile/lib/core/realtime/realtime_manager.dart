import 'dart:async';
import 'dart:developer' as developer;

import 'package:cookie_jar/cookie_jar.dart';
import 'package:letslive_mobile/core/realtime/realtime_channel.dart';
import 'package:letslive_mobile/core/realtime/realtime_frame.dart';

typedef Unsubscribe = void Function();

enum RealtimeStatus {
  /// [RealtimeManager.start] has not been called yet.
  idle,

  /// First connection attempt in flight.
  connecting,
  connected,

  /// Socket lost; a retry is scheduled or in flight.
  reconnecting,

  /// Closed on purpose because the app is in the background.
  paused,

  /// The server rejected the session and one refresh did not fix it. No
  /// further retries until [RealtimeManager.start] is called again (after a
  /// fresh login).
  unauthorized,
}

/// The single shared WebSocket client. Mirrors `web/contexts/realtime-context.tsx`:
/// ref-counted topics, resubscribe on reconnect, `onReconnect` hooks so
/// features can refetch what they missed, backoff 1s doubling to 30s.
///
/// Auth: the realtime service never rejects an upgrade
/// (`backend/realtime/api/server.go` treats a bad/expired ACCESS_TOKEN cookie
/// as an anonymous viewer), so an expired cookie silently yields a socket
/// without the user topic. The manager therefore refreshes BEFORE connecting
/// when the jar holds a REFRESH_TOKEN but no (unexpired) ACCESS_TOKEN, as the
/// web does. A gateway-level rejection (HTTP 401/403 on the upgrade, or close
/// code 1008/4401) is also handled: refresh once, reconnect, and if that fails
/// again before a successful open, stop with [RealtimeStatus.unauthorized].
///
/// Lifecycle: call [start] once the app is up and again after login (also
/// recovers from `unauthorized`); [restart] on any identity change; [pause] /
/// [resume] from the app lifecycle (wired in the shell, Task 11).
class RealtimeManager {
  RealtimeManager({
    required String url,
    required CookieJar cookieJar,
    required this._refreshSession,
    Stream<void>? sessionExpired,
    RealtimeChannelFactory connect = connectIoChannel,
    this._pingInterval = defaultPingInterval,
    this._pongTimeout = defaultPongTimeout,
  }) : _uri = Uri.parse(url),
       _jar = cookieJar,
       _factory = connect {
    _expiredSubscription = sessionExpired?.listen((_) => _onSessionExpired());
  }

  /// Server pings every 30s (`client.PingInterval`,
  /// `backend/realtime/client/client.go:21`); the app-level ping stays below
  /// that so mobile NATs and the gateway idle timeout never see a quiet socket.
  static const Duration defaultPingInterval = Duration(seconds: 25);

  /// Any inbound frame within this window after a ping proves liveness.
  static const Duration defaultPongTimeout = Duration(seconds: 10);

  /// `web/constant/realtime.ts`.
  static const Duration initialReconnectDelay = Duration(seconds: 1);
  static const Duration maxReconnectDelay = Duration(seconds: 30);

  static const Set<int> _unauthorizedHttpStatuses = {401, 403};
  static const Set<int> _unauthorizedCloseCodes = {1008, 4401};

  final Uri _uri;
  final CookieJar _jar;
  final Future<bool> Function() _refreshSession;
  final RealtimeChannelFactory _factory;
  final Duration _pingInterval;
  final Duration _pongTimeout;

  final Map<String, int> _topicCounts = {};
  final Map<String, Set<void Function(Object?)>> _eventHandlers = {};
  final Set<void Function()> _reconnectHandlers = {};
  final StreamController<RealtimeStatus> _statusController =
      StreamController<RealtimeStatus>.broadcast();

  RealtimeStatus _status = RealtimeStatus.idle;
  RealtimeChannel? _channel;
  StreamSubscription<String>? _subscription;
  Timer? _retryTimer;
  Timer? _pingTimer;
  Timer? _pongTimer;
  Duration _delay = initialReconnectDelay;
  // Bumped whenever the current socket/attempt is abandoned; callbacks of
  // older attempts compare and bail out.
  int _generation = 0;
  bool _started = false;
  bool _disposed = false;
  bool _open = false;
  bool _hasOpened = false;
  bool _refreshedSinceOpen = false;
  bool _lastRefreshFailed = false;
  StreamSubscription<void>? _expiredSubscription;

  Stream<RealtimeStatus> get status => _statusController.stream;
  RealtimeStatus get currentStatus => _status;

  /// Opens the socket. No-op while already running; after
  /// [RealtimeStatus.unauthorized] it starts over with a clean slate.
  void start() {
    if (_disposed) return;
    if (_started && _status != RealtimeStatus.unauthorized) return;
    _started = true;
    _refreshedSinceOpen = false;
    _lastRefreshFailed = false;
    _delay = initialReconnectDelay;
    unawaited(_connect());
  }

  /// Drops the socket and reconnects now with the current cookies (login,
  /// logout, account switch). Topics stay subscribed.
  void restart() {
    if (_disposed || !_started) return;
    if (_status == RealtimeStatus.unauthorized) {
      start();
      return;
    }
    _abandonSocket();
    _refreshedSinceOpen = false;
    _lastRefreshFailed = false;
    _delay = initialReconnectDelay;
    if (_status == RealtimeStatus.paused) return;
    unawaited(_connect());
  }

  /// App went to the background: close the socket, keep topic ref-counts.
  void pause() {
    if (_disposed || !_started) return;
    if (_status == RealtimeStatus.paused ||
        _status == RealtimeStatus.unauthorized) {
      return;
    }
    _abandonSocket();
    _setStatus(RealtimeStatus.paused);
  }

  /// App is back: reconnect, resubscribe and fire [onReconnect] handlers.
  void resume() {
    if (_disposed || _status != RealtimeStatus.paused) return;
    _delay = initialReconnectDelay;
    unawaited(_connect());
  }

  /// The server rejected the refresh token: stop everything until
  /// [start]/[restart] (after a fresh login).
  void _onSessionExpired() {
    if (_disposed || !_started) return;
    _abandonSocket();
    _setStatus(RealtimeStatus.unauthorized);
  }

  Future<void> dispose() async {
    if (_disposed) return;
    _disposed = true;
    _abandonSocket();
    await _expiredSubscription?.cancel();
    await _statusController.close();
  }

  /// Joins [topic] (e.g. `room:<id>`) until the returned function runs.
  Unsubscribe subscribe(String topic) {
    final count = (_topicCounts[topic] ?? 0) + 1;
    _topicCounts[topic] = count;
    if (count == 1) _sendTopic('subscribe', topic);

    var active = true;
    return () {
      if (!active) return;
      active = false;
      final remaining = (_topicCounts[topic] ?? 1) - 1;
      if (remaining > 0) {
        _topicCounts[topic] = remaining;
        return;
      }
      _topicCounts.remove(topic);
      _sendTopic('unsubscribe', topic);
    };
  }

  /// Calls [handler] with the event payload for every frame of [type].
  Unsubscribe onEvent(String type, void Function(Object? data) handler) {
    final handlers = _eventHandlers.putIfAbsent(type, () => {});
    handlers.add(handler);
    return () {
      handlers.remove(handler);
      if (handlers.isEmpty && identical(_eventHandlers[type], handlers)) {
        _eventHandlers.remove(type);
      }
    };
  }

  /// Fires after every successful open except the first one: refetch what
  /// was missed while disconnected.
  Unsubscribe onReconnect(void Function() handler) {
    _reconnectHandlers.add(handler);
    return () => _reconnectHandlers.remove(handler);
  }

  // --- connection ---------------------------------------------------------

  Future<void> _connect() async {
    _retryTimer?.cancel();
    final generation = ++_generation;
    _setStatus(
      _hasOpened ? RealtimeStatus.reconnecting : RealtimeStatus.connecting,
    );

    final Map<String, String>? headers;
    try {
      headers = await _buildHeaders();
    } on Exception catch (e) {
      _log('preparing connection failed: $e');
      if (generation == _generation) _scheduleRetry();
      return;
    }
    if (generation != _generation) return;
    if (headers == null) {
      // Logged in but the session could not be refreshed (network blip, or a
      // dead refresh token that `sessionExpired` will report). An anonymous
      // socket would silently miss notifications/DMs, so retry with backoff.
      _scheduleRetry();
      return;
    }

    final RealtimeChannel channel;
    try {
      channel = _factory(_uri, headers);
      _channel = channel;
      await channel.ready;
    } on RealtimeHandshakeException catch (e) {
      if (generation != _generation) return;
      _onDisconnected(handshakeStatus: e.statusCode);
      return;
    } on Exception catch (e) {
      _log('connect failed: $e');
      if (generation == _generation) _onDisconnected();
      return;
    }
    if (generation != _generation) return;

    _subscription = channel.incoming.listen(
      _onFrame,
      onError: (Object _) => _onClosedBy(generation),
      onDone: () => _onClosedBy(generation),
    );
    _onOpen(channel);
  }

  void _onOpen(RealtimeChannel channel) {
    _open = true;
    _delay = initialReconnectDelay;
    _refreshedSinceOpen = false;
    _lastRefreshFailed = false;
    _setStatus(RealtimeStatus.connected);
    // A new socket starts with no topics; restore every one still in use.
    for (final topic in _topicCounts.keys.toList()) {
      channel.send(encodeRealtimeFrame('subscribe', topic: topic));
    }
    if (_hasOpened) {
      for (final handler in List.of(_reconnectHandlers)) {
        _guard('onReconnect', handler);
      }
    }
    _hasOpened = true;
    _pingTimer = Timer.periodic(_pingInterval, (_) => _sendPing());
  }

  void _onClosedBy(int generation) {
    if (generation != _generation) return;
    _onDisconnected(closeCode: _channel?.closeCode);
  }

  /// Socket lost or never opened, and not by our own doing.
  void _onDisconnected({int? handshakeStatus, int? closeCode}) {
    _abandonSocket();
    final unauthorized =
        _unauthorizedHttpStatuses.contains(handshakeStatus) ||
        _unauthorizedCloseCodes.contains(closeCode);
    if (unauthorized) {
      unawaited(_refreshThenReconnect());
    } else {
      _scheduleRetry();
    }
  }

  Future<void> _refreshThenReconnect() async {
    if (_refreshedSinceOpen) {
      // At most one refresh between successful opens. If it succeeded and the
      // server still rejects us, give up; if it failed (transient), back off
      // without refreshing again.
      if (_lastRefreshFailed) {
        _scheduleRetry();
      } else {
        _setStatus(RealtimeStatus.unauthorized);
      }
      return;
    }
    _refreshedSinceOpen = true;
    _setStatus(RealtimeStatus.reconnecting);
    final generation = _generation;
    var refreshed = false;
    try {
      refreshed = await _refreshSession();
    } on Exception catch (e) {
      _log('refresh threw: $e');
    }
    if (generation != _generation) return;
    _lastRefreshFailed = !refreshed;
    if (refreshed) {
      unawaited(_connect());
    } else {
      // Transient or rejected? A rejection arrives via `sessionExpired`,
      // which stops everything; otherwise keep retrying with backoff.
      _scheduleRetry();
    }
  }

  void _scheduleRetry() {
    _setStatus(RealtimeStatus.reconnecting);
    _retryTimer?.cancel();
    _retryTimer = Timer(_delay, () => unawaited(_connect()));
    final doubled = _delay * 2;
    _delay = doubled > maxReconnectDelay ? maxReconnectDelay : doubled;
  }

  /// Invalidates the current attempt, cancels timers and closes the socket.
  void _abandonSocket() {
    _generation++;
    _open = false;
    _retryTimer?.cancel();
    _pingTimer?.cancel();
    _pongTimer?.cancel();
    _retryTimer = _pingTimer = _pongTimer = null;
    final subscription = _subscription;
    _subscription = null;
    if (subscription != null) unawaited(subscription.cancel());
    final channel = _channel;
    _channel = null;
    if (channel != null) unawaited(_closeQuietly(channel));
  }

  Future<void> _closeQuietly(RealtimeChannel channel) async {
    try {
      await channel.close();
    } on Exception catch (e) {
      _log('close failed: $e');
    }
  }

  Future<Map<String, String>?> _buildHeaders() async {
    final httpUri = _uri.replace(
      scheme: _uri.scheme == 'wss' ? 'https' : 'http',
    );
    var cookies = await _jar.loadForRequest(httpUri);
    final hasAccess = cookies.any(
      (c) => c.name == 'ACCESS_TOKEN' && c.value.isNotEmpty,
    );
    final hasRefresh = cookies.any(
      (c) => c.name == 'REFRESH_TOKEN' && c.value.isNotEmpty,
    );
    // The jar drops expired cookies. Without a fresh access cookie a logged-in
    // user's socket would silently be anonymous.
    if (!hasAccess && hasRefresh) {
      if (!await _refreshSession()) return null;
      cookies = await _jar.loadForRequest(httpUri);
    }
    if (cookies.isEmpty) return const {};
    return {'Cookie': cookies.map((c) => '${c.name}=${c.value}').join('; ')};
  }

  // --- traffic ------------------------------------------------------------

  void _sendTopic(String op, String topic) {
    final channel = _channel;
    if (!_open || channel == null) return;
    channel.send(encodeRealtimeFrame(op, topic: topic));
  }

  void _sendPing() {
    final channel = _channel;
    if (!_open || channel == null) return;
    channel.send(encodeRealtimeFrame('ping'));
    final generation = _generation;
    _pongTimer ??= Timer(_pongTimeout, () {
      if (generation != _generation) return;
      _log('no reply to ping, reconnecting');
      _onDisconnected();
    });
  }

  void _onFrame(String raw) {
    // Any inbound frame proves the socket is alive.
    _pongTimer?.cancel();
    _pongTimer = null;
    final frame = parseRealtimeFrame(raw);
    switch (frame) {
      case RealtimeEventFrame(:final type, :final data):
        final handlers = _eventHandlers[type];
        if (handlers == null) return;
        for (final handler in List.of(handlers)) {
          _guard('event $type', () => handler(data));
        }
      case RealtimeErrorFrame(:final code):
        _log('server error: $code');
      case RealtimePongFrame():
        break;
      case null:
        _log('ignored malformed frame');
    }
  }

  void _guard(String what, void Function() action) {
    try {
      action();
    } on Exception catch (e) {
      _log('handler for $what threw: $e');
    }
  }

  void _setStatus(RealtimeStatus next) {
    if (_status == next || _disposed) return;
    _status = next;
    _statusController.add(next);
  }

  void _log(String message) =>
      developer.log(message, name: 'letslive.realtime');
}
