import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/shop/data/gift_repository.dart';
import 'package:letslive_mobile/features/shop/data/shop_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_repository.dart';

import 'harness.dart';

const String _financeDb = 'letslive_finance';

/// Wallet, shop, inventory and gifts against the real dev stack. The wallet
/// has no public way to receive money (deposits are out of scope), so the
/// test user's SPARK balance is seeded in the dev finance database, like
/// other contract tests do for fixtures the API cannot make.
void main() {
  ContractContext? ctx;
  late WalletRepository wallet;
  late ShopRepository shop;
  late GiftRepository gifts;
  late GiftRepository giftsOfB;
  late String userBId;

  Future<String> userId(ContractUser user) async {
    final me = await user.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    return me.data!.id;
  }

  setUpAll(() async {
    final context = ctx = await startContractContext();
    wallet = WalletRepository(context.userA.client);
    shop = ShopRepository(context.userA.client);
    gifts = GiftRepository(context.userA.client);
    giftsOfB = GiftRepository(context.userB.client);
    userBId = await userId(context.userB);

    // Opening the wallet creates the account; then it gets 1000.00 SPARK.
    final overview = await wallet.overview();
    expect(overview.success, isTrue, reason: overview.key);
    final userAId = await userId(context.userA);
    await runSql(
      _financeDb,
      'INSERT INTO account_balances (account_id, currency_code, balance) '
      "SELECT id, 'SPARK', 100000 FROM accounts WHERE owner_id = '$userAId' "
      'ON CONFLICT (account_id, currency_code) '
      'DO UPDATE SET balance = 100000',
    );
  });

  tearDownAll(() async => ctx?.dispose());

  test('currencies, overview and transactions parse', () async {
    final currencies = await wallet.currencies();
    expect(currencies.success, isTrue, reason: currencies.key);
    expect(
      currencies.data!.map((c) => c.code),
      containsAll(['SPARK', 'FLARE']),
    );

    final overview = await wallet.overview();
    expect(overview.success, isTrue, reason: overview.key);
    expect(
      overview.data!.balanceOf('SPARK').format('en'),
      '1,000',
      reason: 'seeded balance',
    );

    final page = await wallet.transactions(0);
    expect(page.success, isTrue, reason: '${page.key}: ${page.message}');
  });

  test('a purchase moves money into the inventory and the ledger', () async {
    final items = (await shop.items()).data!;
    final rose = items.firstWhere((i) => i.name == 'Rose');
    final before = (await wallet.overview()).data!.balanceOf('SPARK');

    final bought = await shop.purchase(rose.id);
    expect(bought.success, isTrue, reason: '${bought.key}: ${bought.message}');

    final after = (await wallet.overview()).data!.balanceOf('SPARK');
    expect(
      (before.units - after.units) ~/ BigInt.from(10).pow(after.scale),
      BigInt.from(rose.price),
    );
    final inventory = (await gifts.inventory(0)).data!;
    expect(inventory.any((i) => i.shopItemId == rose.id), isTrue);
    final latest = (await wallet.transactions(0)).data!.first;
    expect(latest.type, 'purchase');
    expect(latest.netAmount.isNegative, isTrue);
    expect(latest.netAmount.format('en'), '-${rose.price}');
  });

  test('a gift shows as received by B and sent by A', () async {
    final items = (await shop.items()).data!;
    final crown = items.firstWhere((i) => i.name == 'Crown');

    final sent = await gifts.send(crown.id, userBId);
    expect(sent.success, isTrue, reason: '${sent.key}: ${sent.message}');

    final received = await giftsOfB.received(userBId, 0);
    expect(received.success, isTrue, reason: received.key);
    expect(received.data!.any((g) => g.shopItemId == crown.id), isTrue);
    final mine = await gifts.sent(0);
    expect(mine.success, isTrue, reason: mine.key);
    expect(mine.data!.any((g) => g.shopItemId == crown.id), isTrue);
  });

  test('buying more than the balance fails with the web error key', () async {
    final items = (await shop.items()).data!;
    final priciest = items.reduce((a, b) => a.price >= b.price ? a : b);
    // 1000 - Rose - Crown leaves less than the priciest item (Rocket, 1000).
    final response = await shop.purchase(priciest.id);

    expect(response.success, isFalse);
    expect(response.key, 'res_err_insufficient_balance');
  });
}
