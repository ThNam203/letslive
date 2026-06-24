class ShopItem {
  final String id;
  final String name;
  final String? description;
  final String imageUrl;
  final String animationUrl;
  final int price;
  final String createdAt;

  const ShopItem({
    required this.id,
    required this.name,
    this.description,
    required this.imageUrl,
    required this.animationUrl,
    required this.price,
    required this.createdAt,
  });

  factory ShopItem.fromJson(Map<String, dynamic> json) {
    return ShopItem(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      imageUrl: json['imageUrl'] as String,
      animationUrl: json['animationUrl'] as String,
      price: (json['price'] as num).toInt(),
      createdAt: json['createdAt'] as String,
    );
  }
}

class PurchaseResult {
  final String? giftId;
  final String animationUrl;

  const PurchaseResult({
    this.giftId,
    required this.animationUrl,
  });

  factory PurchaseResult.fromJson(Map<String, dynamic> json) {
    return PurchaseResult(
      giftId: json['giftId'] as String?,
      animationUrl: json['animationUrl'] as String,
    );
  }
}
