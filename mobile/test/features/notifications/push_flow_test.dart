import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/push/push_providers.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/core/widgets/back_notice.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/messages/presentation/conversation_screen.dart';
import 'package:letslive_mobile/features/notifications/presentation/notifications_screen.dart';

import '../auth/auth_test_app.dart';
import '../messages/message_fakes.dart';
import '../messages/messages_test_app.dart';
import 'notifications_test_app.dart';

const String _token = 'id:APA91b-token-1';

/// The location of the top screen, pushed ones included.
String _location(TestApp app) => app.container
    .read(appRouterProvider)
    .routerDelegate
    .currentConfiguration
    .last
    .matchedLocation;

FakeNotificationsServer _server() => FakeNotificationsServer()
  ..conversations.add(conversationJson('c-1'))
  ..threads['c-1'] = threadJson('c-1', 2);

void main() {
  notifTest('the device is registered once the user is signed in', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);

    // The test platform is Android and the app language English.
    expect(server.log.where((l) => l.startsWith('register')), [
      'register $_token android en',
    ]);
    expect(push.permissionRequests, 1);
  });

  notifTest('changing the app language re-sends it for the registered token', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    final app = await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
    );

    await app.container
        .read(localeProvider.notifier)
        .setLocale(const Locale('vi'));
    await tester.pumpAndSettle();

    expect(server.log.where((l) => l.startsWith('register')), [
      'register $_token android en',
      'register $_token android vi',
    ]);
  });

  notifTest('logout unregisters token before clearing session', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    final app = await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
    );

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;

    // The server still knew the user when the token was removed, and the
    // removal came before the logout call.
    final unregister = server.log.indexOf('unregister $_token session=true');
    expect(unregister, isNonNegative);
    expect(unregister, lessThan(server.log.indexOf('logout')));
    expect(push.tokensDeleted, 1);
    expect(await app.hasSessionCookie(), isFalse);
    expect(find.byType(LoginScreen), findsOneWidget);
  });

  notifTest('an expired session still drops the local push token', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    final app = await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
    );

    final ended = app.container
        .read(sessionLifecycleProvider)
        .endSession(notifyServer: false);
    await tester.pumpAndSettle();
    await ended;

    expect(server.log, isNot(contains('logout')));
    expect(push.tokensDeleted, 1);
  });

  notifTest('signing in again registers the device for the new account', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    final app = await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
    );
    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    // FCM hands out a new token after deleteToken.
    push.currentToken = 'id:APA91b-token-2';

    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;

    expect(
      server.log.where((l) => l.startsWith('register')).last,
      'register id:APA91b-token-2 android en',
    );
  });

  notifTest('a tap on a push opens the conversation, back returns', (
    tester,
    socket,
    push,
  ) async {
    final app = await pumpNotifications(
      tester,
      _server(),
      socket: socket,
      push: push,
    );

    push.tap({
      'type': 'dm',
      'referenceId': 'c-1',
      'actionUrl': '/messages/c-1',
    });
    await tester.pumpAndSettle();
    expect(find.byType(ConversationScreen), findsOneWidget);
    expect(_location(app), '/messages/c-1');

    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.byType(NotificationsScreen), findsOneWidget);
  });

  notifTest('tap on push for a deleted item lands on a not-found state '
      'with back', (tester, socket, push) async {
    final app = await pumpNotifications(
      tester,
      _server(),
      socket: socket,
      push: push,
    );

    push.tap({
      'type': 'dm',
      'referenceId': 'gone',
      'actionUrl': '/messages/gone',
    });
    await tester.pumpAndSettle();

    expect(find.byType(BackNotice), findsOneWidget);
    expect(find.text('Conversation not found.'), findsOneWidget);
    await tester.tap(find.text('Back'));
    await tester.pumpAndSettle();
    expect(find.byType(BackNotice), findsNothing);
    expect(_location(app), '/notifications');
  });

  notifTest('tap on push for a deleted video lands on a not-found state', (
    tester,
    socket,
    push,
  ) async {
    await pumpNotifications(tester, _server(), socket: socket, push: push);

    push.tap({'type': 'vod_comment', 'actionUrl': '/users/u-1/vods/deleted'});
    await tester.pumpAndSettle();

    expect(find.byType(BackNotice), findsOneWidget);
  });

  notifTest('a push that names nothing the app opens lands on notifications', (
    tester,
    socket,
    push,
  ) async {
    final app = await pumpNotifications(
      tester,
      _server(),
      socket: socket,
      push: push,
      openTab: false,
    );
    expect(_location(app), '/');

    push.tap({'type': 'gift_received', 'actionUrl': 'https://evil.example/x'});
    await tester.pumpAndSettle();

    expect(find.byType(NotificationsScreen), findsOneWidget);
    expect(_location(app), '/notifications');
  });

  notifTest('the tap that launched the app opens after the session loads', (
    tester,
    socket,
    push,
  ) async {
    push.launchData = {'type': 'dm', 'referenceId': 'c-1'};
    final app = await pumpTestApp(
      tester,
      server: _server(),
      signedIn: true,
      settle: false,
      overrides: [
        ...socket.overrides,
        pushMessagingProvider.overrideWithValue(push),
      ],
    );
    // As in `main`: init runs before the session (and the router) is ready.
    await app.container.read(pushServiceProvider).init();
    await tester.pumpAndSettle();

    expect(find.byType(ConversationScreen), findsOneWidget);
    expect(_location(app), '/messages/c-1');
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(_location(app), '/');
  });

  testWidgets('a tap while logged out is dropped, not queued for the login', (
    tester,
  ) async {
    final push = FakePushMessaging()..launchData = {'type': 'dm'};
    final server = FakeNotificationsServer();
    final app = await pumpTestApp(
      tester,
      server: server,
      overrides: [pushMessagingProvider.overrideWithValue(push)],
    );
    await app.container.read(pushServiceProvider).init();
    await tester.pumpAndSettle();

    expect(find.byType(LoginScreen), findsOneWidget);
    expect(server.log.where((l) => l.startsWith('register')), isEmpty);
    expect(push.permissionRequests, 0);
  });
}
