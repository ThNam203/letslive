import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/search/data/search_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

import 'harness.dart';

/// Search endpoints against the real dev stack. A live stream needs an RTMP
/// publisher and a VOD a transcoded upload, so one public live row and two
/// public VOD rows with accented Vietnamese titles are inserted straight
/// into the service databases (removed afterwards; two VODs so a second
/// page has something to show). Searching for it without accents proves the server
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
  String? secondVodId;

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
    Future<String> insertVod(String title) async => (await runSql(
      'letslive_vod',
      'INSERT INTO vods (user_id, title, visibility, duration) '
          "VALUES ('$ownerId', '$title', 'public', 60) RETURNING id",
    )).split('\n').first;
    vodId = await insertVod('Trò chơi vod $marker');
    secondVodId = await insertVod('Trò chơi vod hai $marker');
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
        for (final id in [vodId, secondVodId].nonNulls) {
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

  test('a later page holds the next match, not a repeat of page 0', () async {
    // Exactly the two marker videos match: one per page.
    final query = 'tro choi vod ${marker.toLowerCase()}';
    final first = await search.vods(q: query, limit: 1);
    final second = await search.vods(q: query, page: 1, limit: 1);

    expect(first.success, isTrue, reason: '${first.key}: ${first.message}');
    expect(second.success, isTrue, reason: '${second.key}: ${second.message}');
    expect(first.data, hasLength(1));
    expect(second.data, hasLength(1));
    expect(
      {first.data!.single.id, second.data!.single.id},
      {vodId, secondVodId},
    );
  });
}
