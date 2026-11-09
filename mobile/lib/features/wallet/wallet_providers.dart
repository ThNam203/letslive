/// Riverpod providers of the wallet (web `/wallet/overview` and
/// `/wallet/transactions`; the inventory tab is in `shop_providers.dart`).
///
/// Everything is `autoDispose`, so leaving the signed-in shell on logout drops
/// it and the next account never sees the previous account's balance.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_models.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_repository.dart';

/// Wallet REST calls.
final walletRepositoryProvider = Provider<WalletRepository>(
  (ref) => WalletRepository(ref.watch(apiClientProvider)),
);

/// The signed-in user's balances. Invalidate it to reload (after a purchase
/// or a failed one: the balance may have changed either way).
final walletOverviewProvider = FutureProvider.autoDispose<WalletOverview>(
  (ref) async =>
      unwrapData(await ref.watch(walletRepositoryProvider).overview()),
  retry: noRetry,
);

/// The user's transactions, newest first, paged. The overview shows the first
/// few of them, so both tabs cost one request.
final transactionsProvider =
    AsyncNotifierProvider.autoDispose<
      TransactionsNotifier,
      PagedFeed<WalletTransaction>
    >(TransactionsNotifier.new, retry: noRetry);

/// Loads the transaction pages.
class TransactionsNotifier extends PagedFeedNotifier<WalletTransaction> {
  @override
  int get pageSize => transactionsPageSize;

  @override
  String idOf(WalletTransaction item) => item.id;

  @override
  Future<ApiResponse<List<WalletTransaction>>> fetchPage(int page) =>
      ref.read(walletRepositoryProvider).transactions(page);
}
