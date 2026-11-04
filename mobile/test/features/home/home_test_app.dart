import 'package:dio/dio.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import 'home_fixtures.dart';

/// [FakeAuthServer] plus the home feed endpoints. Each list is what the
/// endpoint answers; a `*Status` other than 200 makes it fail instead.
class FakeHomeServer extends FakeAuthServer {
  List<Map<String, Object?>> live = [];
  List<Map<String, Object?>> vods = [];
  List<Map<String, Object?>> following = [];
  List<Map<String, Object?>> recommended = [];

  /// Public profiles by id; an unknown id answers 404.
  Map<String, Map<String, Object?>> users = {
    'u-1': publicUserJson(id: 'u-1', username: 'streamer'),
  };

  int liveStatus = 200;
  int vodsStatus = 200;
  int followingStatus = 200;

  ResponseBody _list(List<Map<String, Object?>> items, int status) =>
      status == 200
      ? okEnvelope(items)
      : jsonBody({
          'requestId': 'req-err',
          'success': false,
          'code': 1,
          'key': 'res_err_internal_server',
          'message': 'Server down',
        }, status: status);

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    if (o.method == 'GET') {
      switch (o.path) {
        case ApiEndpoints.popularLivestreams:
          return _list(live, liveStatus);
        case ApiEndpoints.popularVods:
          return _list(vods, vodsStatus);
        case ApiEndpoints.following:
          return _list(following, followingStatus);
        case ApiEndpoints.recommendedChannels:
          return _list(recommended, 200);
      }
      if (o.path.startsWith('${ApiEndpoints.user}/') &&
          o.path != ApiEndpoints.userMe) {
        final user = users[o.path.substring(ApiEndpoints.user.length + 1)];
        if (user != null) return okEnvelope(user);
      }
    }
    return super.handle(o);
  }
}

/// Pumps the real app signed in against [server] (home loaded). The screen is
/// tall so every section is built (the feed is a lazy list).
Future<TestApp> pumpHome(WidgetTester tester, FakeHomeServer server) {
  tester.view.physicalSize = const Size(800, 6000);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  return pumpTestApp(tester, server: server, signedIn: true);
}
