import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';

import '../../core/network/fake_adapter.dart';
import '../auth/auth_test_app.dart';
import '../home/home_fixtures.dart';
import '../watch/fake_live_player.dart';
import '../watch/watch_test_app.dart';

const playbackUrl = 'http://localhost:8000/files/vods/vod-1/index.m3u8';

/// A video the way the backend sends it, with a playback URL.
Map<String, Object?> vodDetailJson({
  String id = 'vod-1',
  String userId = 'u-1',
  String title = 'Vod title',
  String? description,
  int duration = 3725,
  String? url = playbackUrl,
  int likeCount = 3,
}) => {
  ...vodJson(id: id, userId: userId, title: title, duration: duration),
  'description': description,
  'playbackUrl': url,
  'likeCount': likeCount,
};

/// A comment the way the backend sends it. `user` is the author unless
/// [author] is null.
Map<String, Object?> commentJson({
  String id = 'c-1',
  String vodId = 'vod-1',
  String userId = 'u-2',
  String username = 'someone',
  String content = 'Nice video',
  String? parentId,
  bool isDeleted = false,
  bool isEdited = false,
  int likeCount = 0,
  int replyCount = 0,
}) => {
  'id': id,
  'vodId': vodId,
  'userId': userId,
  'parentId': parentId,
  'content': content,
  'isDeleted': isDeleted,
  'isEdited': isEdited,
  'likeCount': likeCount,
  'replyCount': replyCount,
  'createdAt': DateTime.now()
      .subtract(const Duration(minutes: 3))
      .toUtc()
      .toIso8601String(),
  'updatedAt': '2026-11-04T00:00:00Z',
  'user': {'id': userId, 'username': username},
};

ResponseBody _fail(int status, String key) => jsonBody({
  'requestId': 'req-err',
  'success': false,
  'code': 1,
  'key': key,
  'message': 'Server said no',
}, status: status);

/// [FakeWatchServer] plus the video and comment endpoints. State lives in
/// the public fields so a test sets it up and reads what the app did.
class FakeVodServer extends FakeWatchServer {
  /// Videos by id; an unknown id answers 404 `res_err_vod_not_found`.
  Map<String, Map<String, Object?>> vodById = {'vod-1': vodDetailJson()};

  /// Public videos of a user, for `GET /vods?userId=`.
  List<Map<String, Object?>> userVods = [];

  /// Status `GET /vods/<id>` answers with (e.g. 503).
  int vodStatus = 200;

  /// When set, `GET /vods/<id>` waits for it before answering.
  Completer<void>? holdVod;

  /// Status `POST /vods/<id>/view` answers with (400 = threshold refused).
  int viewStatus = 200;

  /// Top-level comments of every video, newest first.
  List<Map<String, Object?>> comments = [];

  /// Replies by parent comment id.
  Map<String, List<Map<String, Object?>>> replies = {};

  /// Comment ids the signed-in user liked.
  Set<String> likedIds = {};

  /// The viewer's reaction answered by `GET /vods/<id>/reaction`.
  String? reaction;
  int reactionStatus = 200;

  /// Status of PUT/DELETE reaction and of like/unlike, edit, delete and
  /// create comment calls (e.g. 503 to make them fail).
  int writeStatus = 200;

  /// Status `GET /vod-comments/<id>/replies` answers with.
  int repliesStatus = 200;

  /// Status `GET /vods/<id>/comments` answers with.
  int commentsStatus = 200;

  /// When set, like/unlike wait for it before answering.
  Completer<void>? holdLike;

  /// When set, PUT/DELETE reaction wait for it before answering.
  Completer<void>? holdReaction;

  /// Bodies of the writes, in order: view, reaction, comment calls.
  final List<Map<String, Object?>> viewBodies = [];
  final List<Map<String, Object?>> createdBodies = [];
  final List<Map<String, Object?>> editBodies = [];
  final List<String> deletedIds = [];

  int _created = 0;

  bool _isVodPath(String path, [String suffix = '']) {
    final parts = path.split('/');
    return path.startsWith('${ApiEndpoints.vods}/') &&
        parts.length == (suffix.isEmpty ? 3 : 4) &&
        parts[2] != 'search' &&
        (suffix.isEmpty || parts[3] == suffix);
  }

  String _vodId(RequestOptions o) => o.path.split('/')[2];

  ResponseBody _reactionBody(int likeCount) =>
      okEnvelope({'likeCount': likeCount, 'reaction': reaction});

  Future<ResponseBody?> _handleVod(RequestOptions o) async {
    final method = o.method;
    if (method == 'GET' && o.path == ApiEndpoints.vods) {
      return okEnvelope(userVods);
    }
    if (_isVodPath(o.path)) {
      await holdVod?.future;
      if (vodStatus != 200) return _fail(vodStatus, 'res_err_internal_server');
      final vod = vodById[_vodId(o)];
      return vod == null
          ? _fail(404, 'res_err_vod_not_found')
          : okEnvelope(vod);
    }
    if (_isVodPath(o.path, 'view') && method == 'POST') {
      viewBodies.add(o.data as Map<String, Object?>);
      return viewStatus == 200
          ? okEnvelope(null)
          : _fail(viewStatus, 'res_err_vod_view_threshold');
    }
    if (_isVodPath(o.path, 'reaction')) {
      final likes = (vodById[_vodId(o)]?['likeCount'] as int?) ?? 0;
      if (method == 'GET') {
        return reactionStatus == 200
            ? _reactionBody(likes + (reaction == 'like' ? 1 : 0))
            : _fail(reactionStatus, 'res_err_internal_server');
      }
      await holdReaction?.future;
      if (writeStatus != 200) {
        return _fail(writeStatus, 'res_err_internal_server');
      }
      reaction = method == 'PUT'
          ? (o.data as Map<String, Object?>)['reaction'] as String
          : null;
      return _reactionBody(likes + (reaction == 'like' ? 1 : 0));
    }
    return null;
  }

  Future<ResponseBody?> _handleComments(RequestOptions o) async {
    final method = o.method;
    if (_isVodPath(o.path, 'comments')) {
      if (method == 'GET') {
        if (commentsStatus != 200) {
          return _fail(commentsStatus, 'res_err_internal_server');
        }
        final page = o.queryParameters['page'] as int;
        final limit = o.queryParameters['limit'] as int;
        return okEnvelope(
          comments.skip(page * limit).take(limit).toList(),
          meta: {'total': comments.length},
        );
      }
      final body = o.data as Map<String, Object?>;
      createdBodies.add(body);
      if (writeStatus != 200) {
        return _fail(writeStatus, 'res_err_internal_server');
      }
      final id = 'new-${++_created}';
      final parentId = body['parentId'] as String?;
      final created = commentJson(
        id: id,
        userId: 'u-1',
        content: body['content']! as String,
        parentId: parentId,
      )..remove('user'); // The backend's answer has no author.
      if (parentId == null) {
        comments = [created, ...comments];
      } else {
        replies[parentId] = [...?replies[parentId], created];
      }
      return okEnvelope(created);
    }
    if (!o.path.startsWith('${ApiEndpoints.vodComments}/')) return null;
    final parts = o.path.split('/');
    if (o.path == '${ApiEndpoints.vodComments}/liked-ids') {
      final asked =
          (o.data as Map<String, Object?>)['commentIds']! as List<Object?>;
      return okEnvelope([
        for (final id in asked)
          if (likedIds.contains(id)) id,
      ]);
    }
    final id = parts[2];
    if (parts.length == 4 && parts[3] == 'replies') {
      return repliesStatus == 200
          ? okEnvelope(replies[id] ?? [])
          : _fail(repliesStatus, 'res_err_internal_server');
    }
    if (parts.length == 4 && parts[3] == 'like') {
      await holdLike?.future;
      if (writeStatus != 200) {
        return _fail(writeStatus, 'res_err_internal_server');
      }
      if (method == 'POST') {
        likedIds.add(id);
      } else {
        likedIds.remove(id);
      }
      return okEnvelope(null);
    }
    if (parts.length == 3 && method == 'PATCH') {
      if (writeStatus != 200) {
        return _fail(writeStatus, 'res_err_internal_server');
      }
      final body = o.data as Map<String, Object?>;
      editBodies.add(body);
      return okEnvelope(
        commentJson(id: id, content: body['content']! as String),
      );
    }
    if (parts.length == 3 && method == 'DELETE') {
      if (writeStatus != 200) {
        return _fail(writeStatus, 'res_err_internal_server');
      }
      deletedIds.add(id);
      return okEnvelope(null);
    }
    return null;
  }

  @override
  Future<ResponseBody> handle(RequestOptions o) async =>
      await _handleVod(o) ?? await _handleComments(o) ?? await super.handle(o);
}

/// Opens video `vod-1` of `u-1` again in an app that is already pumped.
Future<void> pumpVodAgain(WidgetTester tester, TestApp app) async {
  unawaited(
    app.container.read(appRouterProvider).push(AppRoutes.vod('u-1', 'vod-1')),
  );
  await tester.pumpAndSettle();
}

/// Pumps the signed-in app and opens video [vodId] of [userId] the way a
/// card does (pushed above Home).
Future<TestApp> pumpVod(
  WidgetTester tester,
  FakeVodServer server,
  FakeLivePlayers players, {
  String userId = 'u-1',
  String vodId = 'vod-1',
  double width = 800,
  double height = 6000,
  double textScale = 1,
  bool settle = true,
  List<Override> overrides = const [],
}) async {
  final app = await pumpWatchHome(
    tester,
    server,
    players,
    width: width,
    height: height,
    textScale: textScale,
    overrides: overrides,
  );
  unawaited(
    app.container.read(appRouterProvider).push(AppRoutes.vod(userId, vodId)),
  );
  if (settle) {
    await tester.pumpAndSettle();
  } else {
    await tester.pump();
  }
  return app;
}
