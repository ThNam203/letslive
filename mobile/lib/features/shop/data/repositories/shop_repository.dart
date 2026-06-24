import '../../../../core/network/api_client.dart';
import '../../../../core/network/api_endpoints.dart';
import '../../../../core/network/api_response.dart';
import '../models/shop_item_model.dart';

class ShopRepository {
  final ApiClient _client;

  ShopRepository(this._client);

  Future<ApiResponse<List<ShopItem>>> getItems() {
    return _client.get(
      ApiEndpoints.financeShopItems,
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => ShopItem.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<ShopItem>> getItemById(String id) {
    return _client.get(
      ApiEndpoints.financeShopItemById(id),
      fromJsonT: (json) => ShopItem.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<PurchaseResult>> purchase({
    required String shopItemId,
    required int quantity,
    String? recipientUserId,
  }) {
    final data = <String, dynamic>{
      'shopItemId': shopItemId,
      'quantity': quantity,
    };
    if (recipientUserId != null) {
      data['recipientUserId'] = recipientUserId;
    }

    return _client.post(
      ApiEndpoints.financeShopPurchase,
      data: data,
      fromJsonT: (json) =>
          PurchaseResult.fromJson(json as Map<String, dynamic>),
    );
  }
}
