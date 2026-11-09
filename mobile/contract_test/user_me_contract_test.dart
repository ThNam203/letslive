import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';

import 'harness.dart';

void main() {
  late ContractContext ctx;

  setUpAll(() async {
    ctx = await startContractContext();
  });

  tearDownAll(() async {
    await ctx.dispose();
  });

  test('contract smoke: GET /user/me returns parseable user', () async {
    final res = await ctx.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );

    expect(res.success, isTrue, reason: '${res.key}: ${res.message}');
    expect(res.failure, isNull);
    final me = res.data;
    expect(me, isNotNull);
    expect(me!.email, ctx.userA.email);
    expect(me.username, ctx.userA.username);
    expect(me.id, isNotEmpty);
  });

  test('contract smoke: two distinct users are authenticated', () async {
    final a = await ctx.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    final b = await ctx.userB.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );

    expect(a.data?.id, isNot(b.data?.id));
    expect(b.data?.email, ctx.userB.email);
  });
}
