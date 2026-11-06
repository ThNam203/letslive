import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/push/device_repository.dart';
import 'package:letslive_mobile/features/notifications/data/notification_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

import 'harness.dart';

const String _userDb = 'letslive_user';

/// Notifications and device registration against the real dev stack. The
/// stack has no public way to create a notification (services create them
/// over the internal API), so the rows are inserted into the user database,
/// like the other contract tests do for fixtures the API cannot make.
void main() {
  ContractContext? ctx;
  late NotificationRepository notifications;
  late DeviceRepository devices;
  late String userId;

  setUpAll(() async {
    final context = ctx = await startContractContext();
    notifications = NotificationRepository(context.userA.client);
    devices = DeviceRepository(context.userA.client);
    final me = await context.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    userId = me.data!.id;
  });

  tearDownAll(() async => ctx?.dispose());

  Future<void> insert(String title, {String? actionUrl}) => runSql(
    _userDb,
    'INSERT INTO notifications (user_id, type, title, message, action_url, '
    "action_label) VALUES ('$userId', 'vod_comment', '$title', 'm', "
    "${actionUrl == null ? 'NULL' : "'$actionUrl'"}, "
    "${actionUrl == null ? 'NULL' : "'View'"})",
  );

  test('list, unread count, mark read, mark all read and delete', () async {
    // New accounts may already have notifications (e.g. a welcome one).
    final baseline = (await notifications.unreadCount()).data!;
    await insert('first');
    await insert('second', actionUrl: '/users/$userId/vods/$userId');

    final list = await notifications.list();
    expect(list.success, isTrue, reason: '${list.key}: ${list.message}');
    final items = list.data!;
    expect(items.map((n) => n.title), containsAll(['first', 'second']));
    final second = items.firstWhere((n) => n.title == 'second');
    expect(second.actionUrl, '/users/$userId/vods/$userId');
    expect(second.actionLabel, 'View');
    expect(second.isRead, isFalse);
    expect(items.first.createdAt.isBefore(DateTime(2100)), isTrue);

    final count = await notifications.unreadCount();
    expect(count.success, isTrue, reason: '${count.key}: ${count.message}');
    expect(count.data, baseline + 2);

    final read = await notifications.markRead(second.id);
    expect(read.success, isTrue, reason: '${read.key}: ${read.message}');
    expect((await notifications.unreadCount()).data, baseline + 1);

    final all = await notifications.markAllRead();
    expect(all.success, isTrue, reason: '${all.key}: ${all.message}');
    expect((await notifications.unreadCount()).data, 0);

    final deleted = await notifications.delete(second.id);
    expect(deleted.success, isTrue, reason: deleted.key);
    final after = await notifications.list();
    expect(after.data!.map((n) => n.id), isNot(contains(second.id)));
  });

  test('a token with a colon registers and is removed, URL-encoded', () async {
    const token = 'id:APA91b-contract-test-token_0123456789';

    final registered = await devices.register(
      token: token,
      platform: 'android',
      locale: 'vi',
    );
    expect(
      registered.success,
      isTrue,
      reason: '${registered.key}: ${registered.message}',
    );
    expect(
      await runSql(
        _userDb,
        "SELECT platform || '/' || locale FROM device_tokens "
        "WHERE token = '$token'",
      ),
      'android/vi',
    );

    // Registering again (a restart) is not an error and stays one row.
    final again = await devices.register(
      token: token,
      platform: 'android',
      locale: 'vi',
    );
    expect(again.success, isTrue);

    final removed = await devices.unregister(token);
    expect(removed.success, isTrue, reason: removed.key);
    expect(
      await runSql(
        _userDb,
        "SELECT count(*) FROM device_tokens WHERE token = '$token'",
      ),
      '0',
    );
  });

  test('an unknown platform is refused', () async {
    final res = await devices.register(
      token: 'id:APA91b-bad-platform',
      platform: 'windows',
      locale: 'en',
    );

    expect(res.success, isFalse);
    expect(res.statusCode, 400);
  });
}
