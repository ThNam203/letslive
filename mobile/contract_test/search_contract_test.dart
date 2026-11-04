import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/search/data/search_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

import 'harness.dart';

/// Search endpoints against the real dev stack. A live stream needs an RTMP
/// publisher and a VOD a transcoded upload, so one public row of each with an
/// accented Vietnamese title is inserted straight into the service databases
/// (removed afterwards). Searching for it without accents proves the server
/// side of "accent-insensitive" and that the real JSON parses with the app's
/// models.
void main() {
  /// Unique per run so rows of earlier runs or other tests never match.
  final marker = 'Zq${DateTime.now().microsecondsSinceEpoch}';
  ContractContext? ctx;
  late String ownerId;
  late SearchRepository search;
  String? livestreamId;
  String? vodId;

  setUpAll(() async {
    final context = ctx = await startContractContext();
    search = SearchRepository(context.userB.client);
    final me = await context.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    ownerId = me.data!.id;
    livestreamId = (await runSql(
      'letslive_livestream',
      "INSERT INTO livestreams (user_id, title) VALUES ('$ownerId', "
          "'Trò chơi live $marker') RETURNING id",
    )).split('\n').first;
    vodId = (await runSql(
      'letslive_vod',
      'INSERT INTO vods (user_id, title, visibility, duration) '
          "VALUES ('$ownerId', 'Trò chơi vod $marker', 'public', 60) "
          'RETURNING id',
    )).split('\n').first;
  });

  tearDownAll(() async {
    try {
      final liveId = livestreamId;
      if (liveId != null) {
        await runSql(
          'letslive_livestream',
          "DELETE FROM livestreams WHERE id = '$liveId'",
        );
      }
    } finally {
      try {
        final id = vodId;
        if (id != null) {
          await runSql('letslive_vod', "DELETE FROM vods WHERE id = '$id'");
        }
      } finally {
        await ctx?.dispose();
      }
    }
  });

  test(
    'GET /vods/search finds an accented title with an accent-less query',
    () async {
      final res = await search.vods(q: 'tro choi vod ${marker.toLowerCase()}');

      expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
      final vod = res.data!.firstWhere((v) => v.id == vodId);
      expect(vod.title, 'Trò chơi vod $marker');
      expect(vod.userId, ownerId);
    },
  );

  test(
    'GET /livestreams/search finds an accented live title without accents',
    () async {
      final res = await search.lives(
        q: 'tro choi live ${marker.toLowerCase()}',
      );

      expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
      final stream = res.data!.firstWhere((s) => s.id == livestreamId);
      expect(stream.title, 'Trò chơi live $marker');
      expect(stream.endedAt, isNull);
    },
  );

  test(
    'GET /users/search finds a channel by username with page and limit',
    () async {
      final res = await search.users(q: ctx!.userA.username, limit: 5);

      expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
      expect(res.data!.map((u) => u.id), contains(ownerId));
      expect(res.data!.length, lessThanOrEqualTo(5));
    },
  );

  test('a search with no match is a successful empty list', () async {
    final res = await search.vods(q: 'qwxyzkjhg plmnbvcxz');

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data ?? const [], isEmpty);
  });

  test('a later page is a list that does not repeat page 0', () async {
    final first = await search.vods(q: 'tro choi', limit: 1);
    final second = await search.vods(q: 'tro choi', page: 1, limit: 1);

    expect(second.success, isTrue, reason: '${second.key}: ${second.message}');
    final firstIds = (first.data ?? const []).map((v) => v.id).toSet();
    expect(second.data ?? const [], hasLength(lessThanOrEqualTo(1)));
    expect(
      (second.data ?? const []).any((v) => firstIds.contains(v.id)),
      isFalse,
    );
  });
}
