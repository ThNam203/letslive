/// Response models of the shop, inventory and gift endpoints, typed after
/// `web/types/shop.ts`.
library;

import 'package:flutter/foundation.dart';
import 'package:letslive_mobile/core/network/json_fields.dart';

/// An item of the shop catalogue (`GET /shop/items`).
@immutable
class ShopItem {
  const ShopItem({
    required this.id,
    required this.name,
    required this.imageUrl,
    required this.price,
    this.description,
  });

  factory ShopItem.fromJson(Map<String, Object?> json) {
    final f = JsonFields('ShopItem', json);
    return ShopItem(
      id: f.string('id'),
      name: f.string('name'),
      description: f.optString('description'),
      imageUrl: f.string('imageUrl'),
      price: f.integer('price'),
    );
  }

  /// `ParseData<List<ShopItem>>` for `ApiClient`.
  static List<ShopItem> parseList(Object? data) =>
      parseJsonList(data, 'ShopItem', ShopItem.fromJson);

  /// The item's id.
  final String id;

  /// Display name.
  final String name;

  /// Short text shown in the shop; null when the item has none.
  final String? description;

  /// Absolute URL of the picture.
  final String imageUrl;

  /// Cost in whole SPARK (an integer, no fraction; web prints it as is).
  final int price;
}

/// An owned item (`GET /user/me/inventory`).
@immutable
class InventoryItem {
  const InventoryItem({
    required this.id,
    required this.shopItemId,
    required this.quantity,
  });

  factory InventoryItem.fromJson(Map<String, Object?> json) {
    final f = JsonFields('InventoryItem', json);
    return InventoryItem(
      id: f.string('id'),
      shopItemId: f.string('shopItemId'),
      quantity: f.integer('quantity'),
    );
  }

  /// `ParseData<List<InventoryItem>>` for `ApiClient`.
  static List<InventoryItem> parseList(Object? data) =>
      parseJsonList(data, 'InventoryItem', InventoryItem.fromJson);

  /// The inventory row's id.
  final String id;

  /// The catalogue item it counts.
  final String shopItemId;

  /// How many the user owns.
  final int quantity;
}

/// A gift sent from one user to another.
@immutable
class Gift {
  const Gift({
    required this.id,
    required this.shopItemId,
    required this.quantity,
    this.message,
  });

  factory Gift.fromJson(Map<String, Object?> json) {
    final f = JsonFields('Gift', json);
    return Gift(
      id: f.string('id'),
      shopItemId: f.string('shopItemId'),
      quantity: f.integer('quantity'),
      message: f.optString('message'),
    );
  }

  /// `ParseData<List<Gift>>` for `ApiClient`.
  static List<Gift> parseList(Object? data) =>
      parseJsonList(data, 'Gift', Gift.fromJson);

  /// The gift's id.
  final String id;

  /// The catalogue item that was sent.
  final String shopItemId;

  /// How many of it.
  final int quantity;

  /// The sender's note; null when there is none.
  final String? message;
}

/// Answer of `POST /shop/purchase`.
@immutable
class PurchaseResult {
  const PurchaseResult({required this.animationUrl});

  factory PurchaseResult.fromJson(Map<String, Object?> json) => PurchaseResult(
    animationUrl: JsonFields('PurchaseResult', json).string('animationUrl'),
  );

  /// `ParseData<PurchaseResult>` for `ApiClient`.
  static PurchaseResult parse(Object? data) =>
      parseJsonObject(data, 'PurchaseResult', PurchaseResult.fromJson);

  /// Picture or animation shown after a gift was sent; may be empty.
  final String animationUrl;
}
