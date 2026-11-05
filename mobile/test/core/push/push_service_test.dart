import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/push/device_repository.dart';
import 'package:letslive_mobile/core/push/push_service.dart';
import 'package:letslive_mobile/core/push/push_tap.dart';

import '../../features/messages/message_fakes.dart';
import '../../features/notifications/notifications_test_app.dart';

/// Records what the service tells the server.
class _FakeDevices implements DeviceRepository {
  final List<String> calls = [];
  bool failRegister = false;
  bool failUnregister = false;
  bool unregisterThrows = false;

  @override
  Future<ApiResponse<Object?>> register({
    required String token,
    required String platform,
    required String locale,
  }) async {
    calls.add('register $token $platform $locale');
    return failRegister ? failed() : ok<Object?>(null);
  }

  @override
  Future<ApiResponse<Object?>> unregister(String token) async {
    calls.add('unregister $token');
    if (unregisterThrows) throw StateError('network gone');
    return failUnregister ? failed() : ok<Object?>(null);
  }
}

PushService _service(FakePushMessaging? messaging, _FakeDevices devices) =>
    PushService(
      messaging: messaging,
      devices: devices,
      platform: 'android',
      locale: () => 'vi',
    );

void main() {
  late FakePushMessaging fcm;
  late _FakeDevices devices;
  late PushService push;

  setUp(() {
    fcm = FakePushMessaging();
    devices = _FakeDevices();
    push = _service(fcm, devices);
    addTearDown(push.dispose);
  });

  test('token registered after login and refreshed on rotation', () async {
    await push.init();
    await push.registerForUser();
    expect(devices.calls, ['register id:APA91b-token-1 android vi']);
    expect(fcm.silenced, isTrue);

    fcm.rotateTo('id:APA91b-token-2');
    await pumpEventQueue();

    expect(devices.calls.last, 'register id:APA91b-token-2 android vi');
  });

  test('a denied permission registers nothing', () async {
    fcm.permission = false;
    await push.init();

    await push.registerForUser();

    expect(devices.calls, isEmpty);
  });

  test('a token that rotates before anyone is signed in is not sent', () async {
    await push.init();

    fcm.rotateTo('id:APA91b-token-2');
    await pumpEventQueue();

    expect(devices.calls, isEmpty);
  });

  test('logout removes the token on the server, then locally', () async {
    await push.init();
    await push.registerForUser();

    await push.unregister();

    expect(devices.calls.last, 'unregister id:APA91b-token-1');
    expect(fcm.tokensDeleted, 1);
    // A rotation after logout is not registered for nobody.
    fcm.rotateTo('id:APA91b-token-3');
    await pumpEventQueue();
    expect(devices.calls, hasLength(2));
  });

  test('the local token is deleted even when the server call fails', () async {
    await push.init();
    await push.registerForUser();

    devices.failUnregister = true;
    await push.unregister();
    expect(fcm.tokensDeleted, 1);

    // A fresh token (FCM hands one out after deleteToken) and a call that
    // throws instead of answering.
    fcm.currentToken = 'id:APA91b-token-2';
    await push.registerForUser();
    devices.unregisterThrows = true;
    await push.unregister();
    expect(devices.calls.last, 'unregister id:APA91b-token-2');
    expect(fcm.tokensDeleted, 2);
  });

  test('a refused registration is retried by the next rotation', () async {
    await push.init();
    devices.failRegister = true;
    await push.registerForUser();

    devices.failRegister = false;
    fcm.rotateTo('id:APA91b-token-2');
    await pumpEventQueue();

    expect(devices.calls.last, 'register id:APA91b-token-2 android vi');
  });

  test('taps are reported, including the one that launched the app', () async {
    fcm.launchData = {'type': 'dm', 'referenceId': 'c-1'};
    final taps = <PushTap>[];
    await push.init();
    // Listening after init: the launch tap was kept.
    push.taps.listen(taps.add);

    fcm.tap({'type': 'vod_comment', 'actionUrl': '/users/u-1/vods/v-1'});
    await pumpEventQueue();

    expect(taps.map((t) => t.type), ['dm', 'vod_comment']);
    expect(taps.first.referenceId, 'c-1');
  });

  test('without Firebase every call does nothing', () async {
    final none = _service(null, devices);
    addTearDown(none.dispose);

    await none.init();
    await none.registerForUser();
    await none.unregister();

    expect(devices.calls, isEmpty);
    expect(none.taps, isA<Stream<PushTap>>());
  });
}
