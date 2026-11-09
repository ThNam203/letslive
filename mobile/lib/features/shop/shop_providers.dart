/// Riverpod providers of the shop, the inventory and gifts (web `/shop`,
/// `/wallet/inventory`, `/users/[userId]/gifts` and the gift modal).
///
/// Everything is `autoDispose`, so leaving the signed-in shell on logout drops
/// it and the next account never sees the previous account's data.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/shop/data/gift_repository.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/data/shop_repository.dart';
import 'package:letslive_mobile/features/wallet/wallet_providers.dart';

/// Shop REST calls.
final shopRepositoryProvider = Provider<ShopRepository>(
  (ref) => ShopRepository(ref.watch(apiClientProvider)),
);

/// Gift and inventory REST calls.
final giftRepositoryProvider = Provider<GiftRepository>(
  (ref) => GiftRepository(ref.watch(apiClientProvider)),
);

/// The shop catalogue. Inventory and gift lists look their items up here by
/// id, as web does.
final shopItemsProvider = FutureProvider.autoDispose<List<ShopItem>>(
  (ref) async => unwrapList(await ref.watch(shopRepositoryProvider).items()),
  retry: noRetry,
);

/// The items the signed-in user owns, paged.
final inventoryProvider =
    AsyncNotifierProvider.autoDispose<
      InventoryNotifier,
      PagedFeed<InventoryItem>
    >(InventoryNotifier.new, retry: noRetry);

/// Loads the inventory pages.
class InventoryNotifier extends PagedFeedNotifier<InventoryItem> {
  @override
  int get pageSize => shopPageSize;

  @override
  String idOf(InventoryItem item) => item.id;

  @override
  Future<ApiResponse<List<InventoryItem>>> fetchPage(int page) =>
      ref.read(giftRepositoryProvider).inventory(page);
}

/// The gifts one user received, newest first, paged.
final receivedGiftsProvider = AsyncNotifierProvider.autoDispose
    .family<ReceivedGiftsNotifier, PagedFeed<Gift>, String>(
      ReceivedGiftsNotifier.new,
      retry: noRetry,
    );

/// Loads the gifts [userId] received.
class ReceivedGiftsNotifier extends PagedFeedNotifier<Gift> {
  ReceivedGiftsNotifier(this.userId);

  /// The recipient.
  final String userId;

  @override
  int get pageSize => shopPageSize;

  @override
  String idOf(Gift item) => item.id;

  @override
  Future<ApiResponse<List<Gift>>> fetchPage(int page) =>
      ref.read(giftRepositoryProvider).received(userId, page);
}

/// Buys items and sends gifts. The state is the id of the item whose request
/// is in flight (null when idle): the screens disable every buy and send
/// button while it is set, and a second call is ignored here too, because a
/// repeated purchase would charge twice. Nothing is retried automatically.
final purchaseProvider =
    NotifierProvider.autoDispose<PurchaseNotifier, String?>(
      PurchaseNotifier.new,
    );

/// See [purchaseProvider].
class PurchaseNotifier extends Notifier<String?> {
  @override
  String? build() => null;

  /// Buys one [itemId] for the signed-in user.
  Future<ApiResponse<PurchaseResult>?> buy(String itemId) => _run(
    itemId,
    () => ref.read(shopRepositoryProvider).purchase(itemId),
    refreshInventory: true,
  );

  /// Buys one [itemId] and sends it to [recipientUserId].
  Future<ApiResponse<PurchaseResult>?> sendGift(
    String itemId,
    String recipientUserId,
  ) => _run(
    itemId,
    () => ref.read(giftRepositoryProvider).send(itemId, recipientUserId),
  );

  /// Runs [request] unless one is in flight (then null). The balance and the
  /// transactions reload afterwards whatever the outcome: a failed or lost
  /// answer may still have moved money. A bought item also reloads the
  /// inventory.
  Future<ApiResponse<PurchaseResult>?> _run(
    String itemId,
    Future<ApiResponse<PurchaseResult>> Function() request, {
    bool refreshInventory = false,
  }) async {
    if (state != null) return null;
    state = itemId;
    final response = await request();
    // Disposed (sheet closed mid-request): the wallet providers are
    // autoDispose too, so the next screen loads fresh data anyway.
    if (!ref.mounted) return response;
    state = null;
    ref.invalidate(walletOverviewProvider);
    ref.invalidate(transactionsProvider);
    if (refreshInventory && response.success) ref.invalidate(inventoryProvider);
    return response;
  }
}
