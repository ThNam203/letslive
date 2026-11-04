import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../home/home_test_app.dart';
import '../search/search_test_app.dart';
import 'fake_live_player.dart';

/// [FakeSearchServer] plus `GET /livestreams?userId=`: [liveByUser] holds
/// the live stream of each user (absent = not live, answered as a success
/// with no data, like the backend). Unknown users on `/user/<id>` answer
/// the backend's 404 `res_err_user_not_found`.
class FakeWatchServer extends FakeSearchServer {
  Map<String, Map<String, Object?>> liveByUser = {};

  /// Status `/livestreams?userId=` answers with (e.g. 503).
  int liveOfStatus = 200;

  /// When set, `/livestreams?userId=` waits for it before answering.
  Completer<void>? holdLiveOf;

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    if (o.method == 'GET' && o.path == ApiEndpoints.livestreams) {
      await holdLiveOf?.future;
      if (liveOfStatus != 200) {
        return jsonBody({
          'requestId': 'req-err',
          'success': false,
          'code': 1,
          'key': 'res_err_internal_server',
          'message': 'Server down',
        }, status: liveOfStatus);
      }
      return okEnvelope(liveByUser[o.queryParameters['userId']]);
    }
    final isProfile =
        o.method == 'GET' &&
        o.path.startsWith('${ApiEndpoints.user}/') &&
        o.path != ApiEndpoints.userMe &&
        !o.path.startsWith('${ApiEndpoints.userMe}/');
    if (isProfile &&
        !users.containsKey(o.path.substring(ApiEndpoints.user.length + 1))) {
      return jsonBody({
        'requestId': 'req-404',
        'success': false,
        'code': 40401,
        'key': 'res_err_user_not_found',
        'message': 'User not found.',
      }, status: 404);
    }
    return super.handle(o);
  }
}

/// The signed-in app on Home with [players] standing in for the native
/// player.
Future<TestApp> pumpWatchHome(
  WidgetTester tester,
  FakeWatchServer server,
  FakeLivePlayers players, {
  double width = 800,
  double height = 6000,
  double textScale = 1,
}) => pumpHome(
  tester,
  server,
  width: width,
  height: height,
  textScale: textScale,
  overrides: [livePlayerFactoryProvider.overrideWithValue(players.create)],
);

/// Pumps the signed-in app and opens the channel of [userId] the way a
/// card does (pushed above Home).
Future<TestApp> pumpChannel(
  WidgetTester tester,
  FakeWatchServer server,
  FakeLivePlayers players,
  String userId, {
  double width = 800,
  double height = 6000,
  double textScale = 1,
  bool settle = true,
}) async {
  final app = await pumpWatchHome(
    tester,
    server,
    players,
    width: width,
    height: height,
    textScale: textScale,
  );
  unawaited(
    app.container.read(appRouterProvider).push(AppRoutes.channel(userId)),
  );
  if (settle) {
    await tester.pumpAndSettle();
  } else {
    await tester.pump();
  }
  return app;
}
