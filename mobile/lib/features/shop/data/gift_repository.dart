import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/data/shop_repository.dart';

/// Gifts and the inventory, mirroring `web/lib/api/gift.ts` plus the gift
/// send of `web/components/user/gift-modal.tsx`.
///
/// Like a purchase, a send moves money and is never retried automatically.
class GiftRepository {
  const GiftRepository(this._client);

  final ApiClient _client;

  /// `GET /user/me/inventory`: the signed-in user's items, [shopPageSize] per
  /// [page] (0-based).
  Future<ApiResponse<List<InventoryItem>>> inventory(int page) => _client.get(
    ApiEndpoints.inventory,
    query: {'page': page, 'limit': shopPageSize},
    parse: InventoryItem.parseList,
  );

  /// `POST /shop/purchase` with a recipient: buys one [itemId] and sends it
  /// to [recipientUserId]. Web sends no message, and neither does the app.
  Future<ApiResponse<PurchaseResult>> send(
    String itemId,
    String recipientUserId,
  ) => _client.post(
    ApiEndpoints.shopPurchase,
    body: {
      'shopItemId': itemId,
      'quantity': 1,
      'recipientUserId': recipientUserId,
    },
    parse: PurchaseResult.parse,
  );

  /// `GET /user/<userId>/gifts/received`: gifts [userId] got, newest first.
  Future<ApiResponse<List<Gift>>> received(String userId, int page) =>
      _client.get(
        '${ApiEndpoints.user}/${Uri.encodeComponent(userId)}/gifts/received',
        query: {'page': page, 'limit': shopPageSize},
        parse: Gift.parseList,
      );

  /// `GET /user/me/gifts/sent`: gifts the signed-in user sent, newest
  /// first. Web has no screen for it; the endpoint exists on the backend.
  Future<ApiResponse<List<Gift>>> sent(int page) => _client.get(
    ApiEndpoints.giftsSent,
    query: {'page': page, 'limit': shopPageSize},
    parse: Gift.parseList,
  );
}
