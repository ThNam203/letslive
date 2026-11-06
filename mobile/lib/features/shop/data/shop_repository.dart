import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';

/// Items per page of inventory and gift lists (the backend allows up to 50;
/// web asks for its default of 20 and shows only that first page).
const int shopPageSize = 20;

/// The shop endpoints, mirroring `web/lib/api/shop.ts`.
///
/// A purchase moves money and is not idempotent: it is sent once, never
/// retried automatically. Callers guard double taps.
class ShopRepository {
  const ShopRepository(this._client);

  final ApiClient _client;

  /// `GET /shop/items`: the whole catalogue.
  Future<ApiResponse<List<ShopItem>>> items() =>
      _client.get(ApiEndpoints.shopItems, parse: ShopItem.parseList);

  /// `POST /shop/purchase`: buys one [itemId] for the signed-in user.
  Future<ApiResponse<PurchaseResult>> purchase(String itemId) => _client.post(
    ApiEndpoints.shopPurchase,
    body: {'shopItemId': itemId, 'quantity': 1},
    parse: PurchaseResult.parse,
  );
}
