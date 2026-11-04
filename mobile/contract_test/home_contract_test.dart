import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/home/data/home_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/user_repository.dart';

import 'harness.dart';

/// Home feed endpoints against the real dev stack. A live stream needs an RTMP
/// publisher and a VOD a transcoded upload, so one public row of each is
/// inserted straight into the service databases (removed afterwards) to prove
/// the real JSON parses with the app's models.
void main() {
  late ContractContext ctx;
  late String ownerId;
  String? livestreamId;
  String? vodId;

  setUpAll(() async {
    ctx = await startContractContext();
    final me = await ctx.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    ownerId = me.data!.id;
    // Live streams list newest first, so this one is on page 0.
    livestreamId = await _insertReturningId(
      'letslive_livestream',
      "INSERT INTO livestreams (user_id, title) VALUES ('$ownerId', "
          "'Contract live Việt 🎉') RETURNING id",
    );
    vodId = await _insertReturningId(
      'letslive_vod',
      // A huge view_count puts it first in the popularity order, so it is
      // always on page 0 whatever else the stack holds.
      'INSERT INTO vods (user_id, title, visibility, duration, view_count) '
          "VALUES ('$ownerId', 'Contract vod', 'public', 3725, 2000000000) "
          'RETURNING id',
    );
    // userA follows userB so the following list is not empty.
    final follow = await ctx.userA.client.post<Object?>(
      '${ApiEndpoints.user}/${await _idOf(ctx.userB)}/follow',
      parse: (data) => data,
    );
    expect(follow.success, isTrue, reason: '${follow.key}: ${follow.message}');
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
        await ctx.dispose();
      }
    }
  });

  test('GET /popular-livestreams parses a live stream', () async {
    final res = await HomeRepository(ctx.userA.client).popularLive();

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    final stream = res.data!.firstWhere((s) => s.id == livestreamId);
    expect(stream.userId, ownerId);
    expect(stream.title, 'Contract live Việt 🎉');
    expect(stream.endedAt, isNull);
    expect(stream.visibility, 'public');
  });

  test('GET /popular-vods parses a video', () async {
    final res = await HomeRepository(ctx.userA.client).popularVods();

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    final vod = res.data!.firstWhere((v) => v.id == vodId);
    expect(vod.userId, ownerId);
    expect(vod.duration, 3725);
    expect(vod.status, 'ready');
  });

  test('GET /popular-vods accepts a later page', () async {
    final res = await HomeRepository(ctx.userA.client).popularVods(page: 1);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data ?? const [], hasLength(lessThanOrEqualTo(20)));
  });

  test('GET /user/<id> parses the owner of a card', () async {
    final res = await UserRepository(ctx.userB.client).byId(ownerId);

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data!.id, ownerId);
    expect(res.data!.username, ctx.userA.username);
  });

  test('GET /user/me/following lists the followed channel', () async {
    final res = await UserRepository(ctx.userA.client).following();

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.data!.map((u) => u.username), contains(ctx.userB.username));
  });

  test('GET /users/recommendations parses channels', () async {
    final res = await UserRepository(ctx.userA.client).recommended();

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
  });
}

Future<String> _idOf(ContractUser user) async {
  final me = await user.client.get<MeUser>(
    ApiEndpoints.userMe,
    parse: MeUser.parse,
  );
  return me.data!.id;
}

/// Runs an `INSERT ... RETURNING id`; psql appends a "INSERT 0 1" status line.
Future<String> _insertReturningId(String database, String sql) async =>
    (await runSql(database, sql)).split('\n').first;
