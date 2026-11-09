import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/notifications/notifications_providers.dart';

import '../auth/auth_test_app.dart';
import 'notifications_test_app.dart';

FakeNotificationsServer _server() => FakeNotificationsServer()
  ..notifications.addAll([
    notificationJson('n-2', title: 'Second', isRead: true),
    notificationJson('n-1', title: 'First'),
  ]);

/// The unread count behind the header bell. The bell itself is covered while
/// the notifications screen is open (it shows the count again on going back,
/// see the bell test below), so this reads the count it shows.
int _badge(WidgetTester tester) =>
    ProviderScope.containerOf(tester.element(find.byType(Scaffold).first))
        .read(notificationBadgeProvider);

/// The number on the visible header bell (0 when hidden).
int _bellLabel(WidgetTester tester) {
  final badge = tester.widget<Badge>(notificationsTabBadge());
  return badge.isLabelVisible ? int.parse((badge.label! as Text).data!) : 0;
}

Map<String, Object?> _pushedJson(String id, {String title = 'Pushed'}) =>
    notificationJson(id, title: title, at: '2026-11-05T09:00:00.000Z');

void main() {
  notifTest('the list shows notifications, marks unread ones and the '
      'badge counts them', (tester, socket, push) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);

    expect(find.text('Notifications'), findsWidgets);
    expect(find.text('First'), findsOneWidget);
    expect(find.text('Second'), findsOneWidget);
    expect(_badge(tester), 1);
    expect(find.byKey(const ValueKey('mark-all-read')), findsOneWidget);
  });

  notifTest('an empty list says so; a failed one offers a retry', (
    tester,
    socket,
    push,
  ) async {
    final server = FakeNotificationsServer()..listNotificationsStatus = 503;
    await pumpNotifications(tester, server, socket: socket, push: push);
    expect(find.text('Try again'), findsOneWidget);

    server
      ..listNotificationsStatus = 200
      ..notifications.clear();
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(find.text('No notifications yet.'), findsOneWidget);
  });

  notifTest('realtime notification increments badge and prepends item', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);
    expect(_badge(tester), 1);

    await socket.push(tester, 'notification.created', _pushedJson('n-3'));
    // The same event twice (a retry) must show and count once: 1 unread
    // before plus the new one.
    await socket.push(tester, 'notification.created', _pushedJson('n-3'));

    expect(_badge(tester), 2);
    expect(find.text('Pushed'), findsOneWidget);
    final firstY = tester.getTopLeft(find.text('Pushed')).dy;
    expect(firstY, lessThan(tester.getTopLeft(find.text('First')).dy));
  });

  notifTest('a notification event the app cannot read refetches instead', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);

    server.notifications.insert(0, notificationJson('n-3', title: 'Pushed'));
    await socket.push(tester, 'notification.created', {'id': 7});
    await tester.pumpAndSettle();

    expect(find.text('Pushed'), findsOneWidget);
    expect(_badge(tester), 2);
  });

  notifTest('mark all read zeroes badge', (tester, socket, push) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);
    expect(_badge(tester), 1);

    await tester.tap(find.byKey(const ValueKey('mark-all-read')));
    await tester.pumpAndSettle();

    expect(server.log, contains('read-all'));
    expect(_badge(tester), 0);
    expect(find.byKey(const ValueKey('mark-all-read')), findsNothing);
  });

  notifTest('the row menu marks one read and deletes', (
    tester,
    socket,
    push,
  ) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);

    await tester.tap(find.byKey(const ValueKey('notification-menu-n-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Mark as read'));
    await tester.pumpAndSettle();
    expect(server.log, contains('read n-1'));
    expect(_badge(tester), 0);

    await tester.tap(find.byKey(const ValueKey('notification-menu-n-2')));
    await tester.pumpAndSettle();
    expect(find.text('Mark as read'), findsNothing);
    await tester.tap(find.text('Delete'));
    await tester.pumpAndSettle();
    expect(server.log, contains('delete n-2'));
    expect(find.text('Second'), findsNothing);
  });

  notifTest('a failed action says why and changes nothing', (
    tester,
    socket,
    push,
  ) async {
    final server = _server()..readStatus = 503;
    await pumpNotifications(tester, server, socket: socket, push: push);

    await tester.tap(find.byKey(const ValueKey('notification-menu-n-1')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Mark as read'));
    await tester.pumpAndSettle();

    expect(find.byType(SnackBar), findsOneWidget);
    expect(_badge(tester), 1);
    expect(find.byKey(const ValueKey('notification-menu-n-1')), findsOneWidget);
  });

  notifTest('tapping a row marks it read and opens its video', (
    tester,
    socket,
    push,
  ) async {
    final server = _server()
      ..notifications[1] = notificationJson(
        'n-1',
        title: 'First',
        actionUrl: '/users/u-1/vods/v-1',
      );
    final app = await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
    );

    await tester.tap(find.text('First'));
    await tester.pumpAndSettle();

    expect(server.log, contains('read n-1'));
    final router = app.container.read(appRouterProvider);
    expect(
      router.routerDelegate.currentConfiguration.last.matchedLocation,
      '/users/u-1/vods/v-1',
    );
  });

  notifTest('a reconnect loads the list and the count again without '
      'duplicates', (tester, socket, push) async {
    final server = _server();
    await pumpNotifications(tester, server, socket: socket, push: push);
    await socket.push(tester, 'notification.created', _pushedJson('n-3'));
    expect(_badge(tester), 2);

    // While the socket was down n-3 was also stored, and n-4 was created.
    server.notifications.insert(0, _pushedJson('n-3'));
    server.notifications.insert(0, _pushedJson('n-4', title: 'Missed'));
    final before = server.log.where((l) => l == 'GET list').length;
    socket.manager.restart();
    await tester.pump();
    socket.channels.last.open();
    await tester.pumpAndSettle();

    expect(server.log.where((l) => l == 'GET list').length, before + 1);
    expect(find.text('Missed'), findsOneWidget);
    expect(find.text('Pushed'), findsOneWidget);
    expect(_badge(tester), 3);
  });

  notifTest('the next account never sees the previous account\'s '
      'notifications', (tester, socket, push) async {
    final server = _server();
    final app = await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
    );
    expect(find.text('First'), findsOneWidget);
    expect(_badge(tester), 1);

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);

    // Account B has one read notification.
    server.notifications
      ..clear()
      ..add(notificationJson('n-9', title: 'For B', isRead: true));
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;
    await openNotificationsTab(tester);

    expect(server.log.where((l) => l == 'GET list'), hasLength(2));
    expect(find.text('For B'), findsOneWidget);
    expect(find.text('First'), findsNothing);
    expect(_badge(tester), 0);
  });

  notifTest('long titles, Vietnamese and emoji do not overflow at 320px '
      'with text scale 2', (tester, socket, push) async {
    final long = 'Một bình luận rất dài không dấu 🎉 ${'x' * 200}';
    final server = FakeNotificationsServer()
      ..notifications.add(
        notificationJson(
          'n-1',
          title: long,
          message: long,
          actionLabel: long,
          actionUrl: '/users/u-1/vods/v-1',
        ),
      );
    await pumpNotifications(
      tester,
      server,
      socket: socket,
      push: push,
      width: 320,
      height: 568,
      textScale: 2,
    );

    expect(find.byKey(const ValueKey('notification-n-1')), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  notifTest('the bar has Home, Messages, Profile; search and the bell are in '
      'every tab header', (tester, socket, push) async {
    await pumpNotifications(
      tester,
      _server(),
      socket: socket,
      push: push,
      openTab: false,
    );

    final labels = tester
        .widgetList<NavigationDestination>(find.byType(NavigationDestination))
        .map((d) => d.label)
        .toList();
    expect(labels, ['Home', 'Messages', 'Profile']);

    for (final tab in labels) {
      await tester.tap(find.byTooltip(tab));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('open-search')), findsOneWidget);
      expect(find.byKey(const ValueKey('open-notifications')), findsOneWidget);
    }
    expect(tester.takeException(), isNull);
  });

  notifTest('the bell opens notifications over the tabs and back returns', (
    tester,
    socket,
    push,
  ) async {
    await pumpNotifications(tester, _server(), socket: socket, push: push);

    expect(find.text('First'), findsOneWidget);
    expect(find.byType(NavigationBar), findsNothing);

    // A notification arrives while the bell is covered.
    await socket.push(tester, 'notification.created', _pushedJson('n-3'));

    await tester.tap(find.byType(BackButton));
    await tester.pumpAndSettle();

    expect(find.byType(NavigationBar), findsOneWidget);
    expect(_bellLabel(tester), 2);
  });
}
