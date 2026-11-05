import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../home/home_fixtures.dart';
import '../profile/profile_test_app.dart';
import '../watch/fake_player.dart';
import '../watch/watch_test_app.dart';

ResponseBody _fail(int status, String key) => jsonBody({
  'requestId': 'req-err',
  'success': false,
  'code': 1,
  'key': key,
  'message': 'Nope',
}, status: status);

/// A chat command the way the backend sends it.
Map<String, Object?> commandJson(
  String id,
  String name, {
  String scope = 'user',
  String response = 'hello there',
  String description = '',
}) => {
  'id': id,
  'scope': scope,
  'ownerId': 'u-1',
  'name': name,
  'response': response,
  'description': description,
  'createdAt': '2026-11-04T00:00:00Z',
};

/// [FakeProfileServer] plus the creator endpoints. State lives in public
/// fields; what the app sent is recorded.
class FakeCreatorServer extends FakeProfileServer {
  FakeCreatorServer() {
    meExtra = {
      'livestreamInformation': {
        'title': 'My stream',
        'description': 'About it',
        'thumbnailUrl': null,
      },
    };
  }

  /// The key `PATCH /user/me/api-key` hands out next.
  int _keys = 0;
  int keyResets = 0;

  /// Status `GET /user/me` answers with (the session only reads it at login).
  int userMeStatus = 200;

  /// Status of the key reset and of the stream information save.
  int streamWriteStatus = 200;

  /// Text fields of the stream information saves, in order.
  final List<Map<String, String>> streamInfoSaves = [];

  /// File field names of those saves (e.g. `thumbnail`), per save.
  final List<List<String>> streamInfoFiles = [];

  /// Chat commands: `user` and `channel` lists, as the server keeps them.
  List<Map<String, Object?>> userCommands = [];
  List<Map<String, Object?>> channelCommands = [];

  /// Bodies of `POST /chat-commands`, and ids of deleted commands.
  final List<Map<String, Object?>> createdCommands = [];
  final List<String> deletedCommands = [];

  /// Status `GET /chat-commands/mine` answers with.
  int commandsStatus = 200;

  /// The creator's videos for `GET /vods/author`.
  List<Map<String, Object?>> myVods = [];

  /// Status `GET /vods/author` answers with.
  int myVodsStatus = 200;

  /// Bodies of `PATCH /vods/<id>`, and ids of deleted videos.
  final List<Map<String, Object?>> vodPatches = [];
  final List<String> deletedVods = [];

  int uploadFileCalls = 0;

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    final path = o.path;
    if (path == ApiEndpoints.userMe &&
        o.method == 'GET' &&
        userMeStatus != 200) {
      return _fail(userMeStatus, 'res_err_internal');
    }
    if (path == ApiEndpoints.streamKey && o.method == 'PATCH') {
      keyResets++;
      if (streamWriteStatus != 200) {
        return _fail(streamWriteStatus, 'res_err_internal');
      }
      return okEnvelope('new-key-${++_keys}');
    }
    if (path == ApiEndpoints.livestreamInformation && o.method == 'PATCH') {
      final form = o.data as FormData;
      final fields = {for (final f in form.fields) f.key: f.value};
      streamInfoSaves.add(fields);
      streamInfoFiles.add([for (final f in form.files) f.key]);
      if (streamWriteStatus != 200) {
        return _fail(streamWriteStatus, 'res_err_internal');
      }
      return okEnvelope({
        'title': fields['title'],
        'description': fields['description'],
        'thumbnailUrl': fields['thumbnailUrl'],
      });
    }
    if (path == ApiEndpoints.uploadFile && o.method == 'POST') {
      return okEnvelope('/files/uploads/${uploadFileCalls++}.jpg');
    }
    if (path.startsWith(ApiEndpoints.chatCommands)) {
      return _commands(o);
    }
    if (path == ApiEndpoints.myVods) {
      if (myVodsStatus != 200) return _fail(myVodsStatus, 'res_err_internal');
      final page = o.queryParameters['page'] as int;
      final limit = o.queryParameters['limit'] as int;
      return okEnvelope(myVods.skip(page * limit).take(limit).toList());
    }
    final isVodWrite =
        path.startsWith('${ApiEndpoints.vods}/') &&
        (o.method == 'PATCH' || o.method == 'DELETE');
    if (isVodWrite) {
      final id = path.split('/').last;
      if (o.method == 'DELETE') {
        deletedVods.add(id);
        myVods.removeWhere((v) => v['id'] == id);
      } else {
        final body = o.data as Map<String, Object?>;
        vodPatches.add(body);
        final index = myVods.indexWhere((v) => v['id'] == id);
        myVods[index] = {
          ...myVods[index],
          'title': body['title'],
          'visibility': body['visibility'],
        };
      }
      return okEnvelope(null);
    }
    return super.handle(o);
  }

  ResponseBody _commands(RequestOptions o) {
    if (o.path == '${ApiEndpoints.chatCommands}/mine') {
      if (commandsStatus != 200) {
        return _fail(commandsStatus, 'res_err_internal');
      }
      return okEnvelope({'user': userCommands, 'channel': channelCommands});
    }
    if (o.method == 'POST') {
      final body = o.data as Map<String, Object?>;
      createdCommands.add(body);
      final name = body['name']! as String;
      final taken = [
        ...userCommands,
        ...channelCommands,
      ].any((c) => c['name'] == name);
      if (taken) return _fail(409, 'res_err_conflict');
      final created = commandJson(
        'c-${createdCommands.length}',
        name,
        scope: body['scope']! as String,
        response: body['response']! as String,
        description: body['description']! as String,
      );
      (body['scope'] == 'user' ? userCommands : channelCommands).add(created);
      return okEnvelope(created);
    }
    final id = o.path.split('/').last;
    if (o.method == 'DELETE') {
      deletedCommands.add(id);
      userCommands.removeWhere((c) => c['id'] == id);
      channelCommands.removeWhere((c) => c['id'] == id);
      return okEnvelope(null);
    }
    final body = o.data as Map<String, Object?>;
    for (final list in [userCommands, channelCommands]) {
      final index = list.indexWhere((c) => c['id'] == id);
      if (index >= 0) {
        list[index] = {...list[index], ...body};
        return okEnvelope(list[index]);
      }
    }
    return _fail(404, 'res_err_not_found');
  }
}

/// The signed-in app on a pushed [route], with [extra] provider overrides on
/// top of the [device] ones.
Future<TestApp> pumpCreator(
  WidgetTester tester,
  FakeCreatorServer server,
  String route, {
  FakeDevice? device,
  List<Override> extra = const [],
  double width = 800,
  double height = 6000,
  double textScale = 1,
}) async {
  final app = await pumpWatchHome(
    tester,
    server,
    FakePlayers(),
    width: width,
    height: height,
    textScale: textScale,
    overrides: [...(device ?? FakeDevice()).overrides, ...extra],
  );
  unawaited(app.container.read(appRouterProvider).push(route));
  await tester.pumpAndSettle();
  return app;
}

/// A creator video the way `GET /vods/author` sends it.
Map<String, Object?> myVodJson(
  String id,
  String title, {
  String status = 'ready',
  String visibility = 'public',
}) => {
  ...vodJson(id: id, userId: 'u-1', title: title),
  'status': status,
  'visibility': visibility,
};
