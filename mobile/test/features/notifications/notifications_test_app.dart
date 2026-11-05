import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/push/push_messaging.dart';
import 'package:letslive_mobile/core/push/push_providers.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../messages/messages_test_app.dart';

/// A notification as the backend serializes it (`domains.Notification`).
Map<String, Object?> notificationJson(
  String id, {
  String type = 'vod_comment',
  String title = 'New comment on your video',
  String message = 'bob commented',
  bool isRead = false,
  String? actionUrl,
  String? actionLabel,
  String? referenceId,
  String at = '2026-11-04T10:00:00.000Z',
}) => {
  'id': id,
  'userId': 'u-1',
  'type': type,
  'title': title,
  'message': message,
  'actionUrl': actionUrl,
  'actionLabel': actionLabel,
  'referenceId': referenceId,
  'isRead': isRead,
  'createdAt': at,
};

/// [FakeMessagesServer] plus the notification and device endpoints, over the
/// in-memory [notifications] (newest first). Everything the app sends is
/// recorded in [log], in order, so tests can check what came first.
class FakeNotificationsServer extends FakeMessagesServer {
  final List<Map<String, Object?>> notifications = [];

  /// Calls in order: `GET list`, `read <id>`, `read-all`, `delete <id>`,
  /// `register <token> <platform> <locale>`, `logout` and
  /// `unregister <token> session=<bool>` (whether the user was still signed
  /// in for it).
  final List<String> log = [];

  /// Status of `GET /user/me/notifications` (e.g. 503).
  int listNotificationsStatus = 200;

  /// Status of `PATCH .../<id>/read` (e.g. 503).
  int readStatus = 200;

  /// Unread count answered by `GET .../unread-count`; null counts the
  /// unread items of [notifications].
  int? unreadOverride;

  int get unreadCount =>
      unreadOverride ?? notifications.where((n) => n['isRead'] == false).length;

  ResponseBody _failure(int status) => jsonBody({
    'requestId': 'req-err',
    'success': false,
    'code': 1,
    'key': 'res_err_internal_server',
    'message': 'Server down',
  }, status: status);

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    final path = o.path;
    if (path == ApiEndpoints.logout) log.add('logout');
    if (path == ApiEndpoints.userDevices) {
      final body = o.data as Map<String, Object?>;
      log.add(
        'register ${body['token']} ${body['platform']} ${body['locale']}',
      );
      return okEnvelope({'id': 'd-1'});
    }
    if (path.startsWith('${ApiEndpoints.userDevices}/')) {
      final token = Uri.decodeComponent(
        path.substring(ApiEndpoints.userDevices.length + 1),
      );
      final hasSession = '${o.headers['cookie'] ?? ''}'.contains(
        'ACCESS_TOKEN=access',
      );
      log.add('unregister $token session=$hasSession');
      return okEnvelope(null);
    }
    if (!path.startsWith(ApiEndpoints.notifications)) return super.handle(o);
    final rest = path.substring(ApiEndpoints.notifications.length);
    switch ((o.method, rest)) {
      case ('GET', ''):
        log.add('GET list');
        if (listNotificationsStatus != 200) {
          return _failure(listNotificationsStatus);
        }
        final page = o.queryParameters['page'] as int;
        return okEnvelope(notifications.skip(page * 20).take(20).toList());
      case ('GET', '/unread-count'):
        return okEnvelope({'count': unreadCount});
      case ('PATCH', '/read-all'):
        log.add('read-all');
        for (final n in notifications) {
          n['isRead'] = true;
        }
        return okEnvelope(null);
    }
    final parts = rest.split('/'); // ['', id, 'read'?]
    final id = parts[1];
    if (o.method == 'PATCH' && parts.length == 3 && parts[2] == 'read') {
      log.add('read $id');
      if (readStatus != 200) return _failure(readStatus);
      for (final n in notifications.where((n) => n['id'] == id)) {
        n['isRead'] = true;
      }
      return okEnvelope(null);
    }
    if (o.method == 'DELETE') {
      log.add('delete $id');
      notifications.removeWhere((n) => n['id'] == id);
      return okEnvelope(null);
    }
    return _failure(404);
  }
}

/// [PushMessaging] the test controls.
class FakePushMessaging implements PushMessaging {
  /// What the permission prompt answers.
  bool permission = true;

  /// The FCM token; null means FCM has none.
  String? currentToken = 'id:APA91b-token-1';

  /// Data of the push that launched the app; set before the app is pumped.
  Map<String, Object?>? launchData;

  bool silenced = false;
  int permissionRequests = 0;
  int tokensDeleted = 0;
  final StreamController<String> _refreshes = StreamController.broadcast();
  final StreamController<Map<String, Object?>> _taps =
      StreamController.broadcast();

  /// FCM rotates the token.
  void rotateTo(String token) {
    currentToken = token;
    _refreshes.add(token);
  }

  /// The user taps a push while the app is in the background.
  void tap(Map<String, Object?> data) => _taps.add(data);

  @override
  Future<void> silenceForeground() async => silenced = true;

  @override
  Future<bool> requestPermission() async {
    permissionRequests++;
    return permission;
  }

  @override
  Future<String?> token() async => currentToken;

  @override
  Stream<String> get tokenRefreshes => _refreshes.stream;

  @override
  Future<void> deleteToken() async {
    tokensDeleted++;
    currentToken = null;
  }

  @override
  Future<Map<String, Object?>?> launchTapData() async => launchData;

  @override
  Stream<Map<String, Object?>> get tapData => _taps.stream;
}

/// Pumps the signed-in app on the Notifications tab against [server] with
/// [socket] as the realtime connection and [push] as FCM. Every service is
/// initialised the way `main` does.
Future<TestApp> pumpNotifications(
  WidgetTester tester,
  FakeNotificationsServer server, {
  required FakeSocket socket,
  required FakePushMessaging push,
  double width = 800,
  double height = 1200,
  double textScale = 1,
  bool openTab = true,
}) async {
  tester.view.physicalSize = Size(width, height);
  tester.view.devicePixelRatio = 1;
  tester.platformDispatcher.textScaleFactorTestValue = textScale;
  addTearDown(tester.view.reset);
  addTearDown(tester.platformDispatcher.clearAllTestValues);
  final app = await pumpTestApp(
    tester,
    server: server,
    signedIn: true,
    overrides: [
      ...socket.overrides,
      pushMessagingProvider.overrideWithValue(push),
    ],
  );
  await app.container.read(pushServiceProvider).init();
  await socket.connect(tester);
  if (openTab) await openNotificationsTab(tester);
  return app;
}

Future<void> openNotificationsTab(WidgetTester tester) async {
  await tester.tap(
    find.descendant(
      of: find.byType(NavigationBar),
      matching: find.byIcon(Icons.notifications_none),
    ),
  );
  await tester.pumpAndSettle();
}

/// The badge on the Notifications tab icon.
Finder notificationsTabBadge() => find.ancestor(
  of: find.descendant(
    of: find.byType(NavigationBar),
    matching: find.byIcon(Icons.notifications_none),
  ),
  matching: find.byType(Badge),
);

/// A widget test with a realtime socket and FCM fakes. The socket is
/// disposed at the end of the test body: its ping timer must be stopped
/// before the framework checks for pending timers (a tear-down is too late).
void notifTest(
  String description,
  Future<void> Function(
    WidgetTester tester,
    FakeSocket socket,
    FakePushMessaging push,
  )
  body,
) => testWidgets(description, (tester) async {
  final socket = FakeSocket();
  try {
    await body(tester, socket, FakePushMessaging());
  } finally {
    await socket.dispose();
  }
});
