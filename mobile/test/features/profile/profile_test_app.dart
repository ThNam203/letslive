import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/l10n/locale_provider.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';
import 'package:letslive_mobile/features/settings/settings_providers.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../home/home_fixtures.dart';
import '../vod/vod_test_app.dart';
import '../watch/fake_player.dart';
import '../watch/watch_test_app.dart';

/// The picture URL the fake server answers to an upload.
const uploadedUrl = 'http://localhost:8000/files/avatars/new.jpg';

ResponseBody _fail(int status, String key, [String message = 'Nope']) =>
    jsonBody({
      'requestId': 'req-err',
      'success': false,
      'code': 1,
      'key': key,
      'message': message,
    }, status: status);

/// [FakeVodServer] plus what profiles need: `u-2` ("other") to follow, a
/// stateful `/user/me` that `PUT` changes, picture uploads, password change
/// and paged channel videos. Settable fields drive the failures; the lists
/// record what the app sent.
class FakeProfileServer extends FakeVodServer {
  FakeProfileServer() {
    users['u-2'] = publicUserJson(id: 'u-2', username: 'other');
    users['u-1'] = publicUserJson(id: 'u-1', username: 'abcdef');
  }

  /// Videos answered by `GET /vods?userId=` (any user), newest first.
  List<Map<String, Object?>> channelVods = [];

  /// Status `GET /vods` answers with.
  int channelVodsStatus = 200;

  /// Page index (0-based) of `GET /vods` that answers 503, or null.
  int? failChannelVodsPage;

  /// Extra fields of `/user/me` (bio, pictures, links, authProvider...).
  Map<String, Object?> meExtra = {};

  /// Bodies of `PUT /user/me`, in order.
  final List<Map<String, Object?>> profileUpdates = [];

  /// Status `PUT /user/me` answers with when not 200 (a server failure).
  int profileUpdateStatus = 200;

  /// Status of follow/unfollow (e.g. 503).
  int followStatus = 200;

  /// When set, follow/unfollow wait for it before answering.
  Completer<void>? holdFollow;

  /// `follow:<id>` / `unfollow:<id>` per call, in order.
  final List<String> followCalls = [];

  /// Status of the picture uploads, and the error key when not 200.
  int uploadStatus = 200;
  String uploadKey = 'res_err_invalid_image';

  /// When set, uploads report half progress, then wait for it.
  Completer<void>? holdUpload;

  /// Status of `PATCH /auth/password` (204 = changed) and its error key.
  int passwordStatus = 204;
  String passwordKey = 'res_err_password_not_match';

  /// Bodies of `PATCH /auth/password`, in order.
  final List<Map<String, Object?>> passwordBodies = [];

  /// Id of the signed-in user answered by `/user/me`.
  String meId = 'u-1';

  ResponseBody _me(RequestOptions o) {
    if (!'${o.headers['cookie'] ?? ''}'.contains('ACCESS_TOKEN=access')) {
      return _fail(401, 'res_err_unauthorized');
    }
    return okEnvelope({
      'id': meId,
      'username': username,
      'email': 'user@example.com',
      'status': 'normal',
      'authProvider': 'local',
      'createdAt': '2026-11-04T00:00:00Z',
      'followerCount': 0,
      'streamAPIKey': 'stream-key',
      ...meExtra,
    });
  }

  Future<ResponseBody?> _handleProfile(RequestOptions o) async {
    final path = o.path;
    if (path == ApiEndpoints.userMe && o.method == 'GET') return _me(o);
    if (path == ApiEndpoints.userMe && o.method == 'PUT') {
      final body = Map<String, Object?>.of(o.data as Map<String, Object?>);
      profileUpdates.add(body);
      if (profileUpdateStatus != 200) {
        return _fail(profileUpdateStatus, 'res_err_internal_server');
      }
      if (body['username'] == 'taken_name') {
        return _fail(409, 'res_err_username_taken', 'Username is taken');
      }
      if (body['username'] case final String name) username = name;
      meExtra = {
        ...meExtra,
        if (body['bio'] case final String bio) 'bio': bio,
        if (body['phoneNumber'] case final String phone) 'phoneNumber': phone,
        if (body['status'] case final String status) 'status': status,
        if (body['locale'] case final String locale) 'locale': locale,
        if (body['socialMediaLinks'] case final Map<String, Object?> links)
          'socialMediaLinks': {
            ...?(meExtra['socialMediaLinks'] as Map<String, Object?>?),
            ...links,
          }..removeWhere((_, url) => url == ''),
      };
      return _me(o);
    }
    if ((path == ApiEndpoints.profilePicture ||
            path == ApiEndpoints.backgroundPicture) &&
        o.method == 'PATCH') {
      o.onSendProgress?.call(50, 100);
      await holdUpload?.future;
      return uploadStatus == 200
          ? okEnvelope(uploadedUrl)
          : _fail(uploadStatus, uploadKey);
    }
    if (path == ApiEndpoints.changePassword && o.method == 'PATCH') {
      final body = o.data as Map<String, Object?>;
      passwordBodies.add(body);
      if (passwordStatus != 204) return _fail(passwordStatus, passwordKey);
      if (body['oldPassword'] != testPassword) {
        return _fail(400, 'res_err_password_not_match');
      }
      return ResponseBody.fromString('', 204);
    }
    return null;
  }

  Future<ResponseBody?> _handleFollow(RequestOptions o) async {
    final parts = o.path.split('/');
    final isFollow =
        parts.length == 4 &&
        '/${parts[1]}' == ApiEndpoints.user &&
        (parts[3] == 'follow' && o.method == 'POST' ||
            parts[3] == 'unfollow' && o.method == 'DELETE');
    if (!isFollow) return null;
    final id = parts[2];
    followCalls.add('${parts[3]}:$id');
    await holdFollow?.future;
    if (followStatus != 200) {
      return _fail(followStatus, 'res_err_internal_server');
    }
    final user = users[id];
    if (user != null) {
      final follow = parts[3] == 'follow';
      users[id] = {
        ...user,
        'isFollowing': follow,
        'followerCount': (user['followerCount']! as int) + (follow ? 1 : -1),
      };
    }
    return okEnvelope(null);
  }

  ResponseBody? _handleChannelVods(RequestOptions o) {
    if (o.method != 'GET' || o.path != ApiEndpoints.vods) return null;
    final page = o.queryParameters['page'] as int;
    final limit = o.queryParameters['limit'] as int;
    if (channelVodsStatus != 200 || page == failChannelVodsPage) {
      return _fail(503, 'res_err_internal_server');
    }
    return okEnvelope(channelVods.skip(page * limit).take(limit).toList());
  }

  @override
  Future<ResponseBody> handle(RequestOptions o) async =>
      await _handleProfile(o) ??
      await _handleFollow(o) ??
      _handleChannelVods(o) ??
      await super.handle(o);
}

/// A device stand-in: the photo library returns [next] (null = cancelled),
/// and every image decodes to [size].
class FakeDevice {
  /// The device language (`vi` for the Vietnamese cases).
  Locale locale = const Locale('en');

  PickedImage? next;
  ({int width, int height})? size = (width: 400, height: 400);

  /// The URLs the app asked to open, and whether opening works.
  final List<Uri> opened = [];
  bool canOpen = true;

  List<Override> get overrides => [
    deviceLocaleProvider.overrideWithValue(locale),
    imagePickProvider.overrideWithValue(() async => next),
    imageSizeReaderProvider.overrideWithValue((_) async => size),
    urlOpenerProvider.overrideWithValue((url) async {
      opened.add(url);
      return canOpen;
    }),
  ];
}

/// A valid 1x1 PNG, so previews can decode it.
final Uint8List _tinyPng = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA'
  '60e6kgAAAABJRU5ErkJggg==',
);

/// A picked image; [bytes] pads it to that length when given (size checks).
PickedImage pickedImage({
  String name = 'photo.jpg',
  String? mimeType = 'image/jpeg',
  int? bytes,
}) => PickedImage(
  name: name,
  bytes: bytes == null ? _tinyPng : Uint8List(bytes),
  mimeType: mimeType,
);

/// The signed-in app on a pushed [route] (a channel, a settings page...).
Future<TestApp> pumpRoute(
  WidgetTester tester,
  FakeProfileServer server,
  String route, {
  FakeDevice? device,
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
    overrides: (device ?? FakeDevice()).overrides,
  );
  unawaited(app.container.read(appRouterProvider).push(route));
  await tester.pumpAndSettle();
  return app;
}

/// A list of [count] public videos of `u-2`, newest first.
List<Map<String, Object?>> vodsOf(int count) => [
  for (var i = 0; i < count; i++)
    vodJson(id: 'v-$i', userId: 'u-2', title: 'Video $i'),
];

/// A full-screen text field finder by key string.
Finder field(String key) => find.byKey(ValueKey(key));
