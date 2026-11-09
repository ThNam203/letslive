import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/app_config.dart';

import '../auth/auth_test_app.dart';
import '../watch/fake_player.dart';
import 'vod_test_app.dart';

final _copyLink = find.byKey(const ValueKey('vod-copy-link'));

/// The dev config of the widget tests plus the web app's address.
final _configWithWeb = AppConfig.fromEnv(const {
  'BACKEND_PROTOCOL': 'http',
  'BACKEND_WS_PROTOCOL': 'ws',
  'BACKEND_HOST': 'localhost',
  'BACKEND_PORT': '8000',
  'FLAVOR': 'dev',
  'TURNSTILE_DEV_TOKEN': 'dev-token',
  'WEB_BASE_URL': 'http://localhost:3000/',
});

/// Records `Clipboard.setData` calls on the platform channel; [fail] makes
/// them throw like a device that refuses.
List<String> _mockClipboard(WidgetTester tester, {bool fail = false}) {
  final copied = <String>[];
  tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
    SystemChannels.platform,
    (call) async {
      if (call.method == 'Clipboard.setData') {
        if (fail) throw PlatformException(code: 'unavailable');
        copied.add(
          (call.arguments as Map<Object?, Object?>)['text']! as String,
        );
      }
      return null;
    },
  );
  addTearDown(
    () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      SystemChannels.platform,
      null,
    ),
  );
  return copied;
}

/// The copy-link button next to like/dislike.
void main() {
  testWidgets('copies the web link of the video and says so', (tester) async {
    final copied = _mockClipboard(tester);
    await pumpVod(
      tester,
      FakeVodServer(),
      FakePlayers(),
      config: _configWithWeb,
    );

    await tester.tap(_copyLink);
    await tester.pump();

    // The path of the video itself (web `/users/<owner>/vods/<id>`), once
    // joined to the origin without a double slash.
    expect(copied, ['http://localhost:3000/users/u-1/vods/vod-1']);
    expect(find.text('Link copied'), findsOneWidget);
  });

  testWidgets('a device that refuses the copy gets an error message', (
    tester,
  ) async {
    _mockClipboard(tester, fail: true);
    await pumpVod(
      tester,
      FakeVodServer(),
      FakePlayers(),
      config: _configWithWeb,
    );

    await tester.tap(_copyLink);
    await tester.pump();

    expect(find.text('Couldn\'t copy the link'), findsOneWidget);
    expect(find.text('Link copied'), findsNothing);
  });

  testWidgets('without WEB_BASE_URL the button is hidden', (tester) async {
    expect(testConfig.webBaseUrl, isNull);
    await pumpVod(tester, FakeVodServer(), FakePlayers());

    expect(_copyLink, findsNothing);
    expect(find.byKey(const ValueKey('vod-like')), findsOneWidget);
  });
}
