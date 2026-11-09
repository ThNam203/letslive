import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/user_repository.dart';
import 'package:letslive_mobile/features/watch/data/watch_repository.dart';

import 'harness.dart';

/// The channel screen's lookups against the real dev stack. A live stream
/// needs an RTMP publisher, so one row is inserted straight into the
/// livestream database (removed afterwards); ending it the way the backend
/// does (`ended_at` set) must make the user offline again.
void main() {
  ContractContext? ctx;
  late String ownerId;
  late WatchRepository watch;
  String? livestreamId;

  setUpAll(() async {
    final context = ctx = await startContractContext();
    watch = WatchRepository(context.userB.client);
    final me = await context.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    ownerId = me.data!.id;
  });

  tearDownAll(() async {
    try {
      final id = livestreamId;
      if (id != null) {
        await runSql(
          'letslive_livestream',
          "DELETE FROM livestreams WHERE id = '$id'",
        );
      }
    } finally {
      await ctx?.dispose();
    }
  });

  test(
    'GET /livestreams?userId= is a success without data when offline',
    () async {
      final res = await watch.liveOf(ownerId);

      expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
      expect(res.data, isNull);
    },
  );

  test(
    'GET /livestreams?userId= parses the live stream, then none once ended',
    () async {
      livestreamId = (await runSql(
        'letslive_livestream',
        "INSERT INTO livestreams (user_id, title) VALUES ('$ownerId', "
            "'Contract channel Việt 🎉') RETURNING id",
      )).split('\n').first;

      final live = await watch.liveOf(ownerId);
      expect(live.success, isTrue, reason: '${live.key}: ${live.message}');
      expect(live.data!.id, livestreamId);
      expect(live.data!.userId, ownerId);
      expect(live.data!.title, 'Contract channel Việt 🎉');
      expect(live.data!.endedAt, isNull);

      await runSql(
        'letslive_livestream',
        "UPDATE livestreams SET ended_at = now() WHERE id = '$livestreamId'",
      );
      final ended = await watch.liveOf(ownerId);
      expect(ended.success, isTrue, reason: '${ended.key}: ${ended.message}');
      expect(ended.data, isNull);
    },
  );

  test(
    'GET /user/<id> answers 404 for an unknown user (not-found state)',
    () async {
      final res = await UserRepository(ctx!.userB.client)
          .byId('00000000-0000-4000-8000-000000000000');

      expect(res.success, isFalse);
      expect(res.statusCode, 404);
      expect(res.key, 'res_err_user_not_found');
    },
  );

  test(
    'GET /user/<id> answers 400 for a malformed id (not-found state)',
    () async {
      final res = await UserRepository(ctx!.userB.client).byId('not-a-uuid');

      expect(res.success, isFalse);
      expect(res.statusCode, 400);
    },
  );
}
