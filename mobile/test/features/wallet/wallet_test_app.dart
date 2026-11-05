import 'dart:async';

import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';

import '../../core/network/fake_adapter.dart';
import '../profile/profile_test_app.dart';

/// A shop item as the finance service serializes it.
Map<String, Object?> shopItemJson(
  String id, {
  String name = 'Rose',
  int price = 10,
  String? description,
}) => {
  'id': id,
  'name': name,
  'description': description,
  'imageUrl': 'http://localhost:8000/files/shop/$id.png',
  'animationUrl': 'http://localhost:8000/files/shop/$id.gif',
  'price': price,
  'currencyCode': 'SPARK',
  'createdAt': '2026-11-04T00:00:00Z',
};

/// A transaction with one ledger entry of the user's account.
Map<String, Object?> transactionJson(
  String id, {
  String type = 'purchase',
  String status = 'completed',
  String amount = '-10.00',
  String currency = 'SPARK',
}) => {
  'id': id,
  'type': type,
  'status': status,
  'reference': null,
  'actorId': 'u-1',
  'metadata': null,
  'createdAt': '2026-11-04T10:00:00Z',
  'entries': [
    {
      'id': 'e-$id',
      'transactionId': id,
      'accountId': 'a-1',
      'currencyCode': currency,
      'amount': amount,
      'createdAt': '2026-11-04T10:00:00Z',
    },
  ],
};

/// [FakeProfileServer] plus the wallet, shop, inventory and gift endpoints.
/// Settable fields drive failures; [purchaseBodies] records what the app sent.
/// A purchase without recipient adds the item to [inventory].
class FakeWalletServer extends FakeProfileServer {
  /// SPARK balance in minor units (cents), answered as `"<n>.<cc>"`.
  int sparkCents = 10000;

  /// FLARE balance string; null leaves the currency out like a new wallet.
  String? flareBalance;

  List<Map<String, Object?>> transactions = [];
  List<Map<String, Object?>> shopItems = [
    shopItemJson('i-1', name: 'Rose', price: 10, description: 'A red rose'),
    shopItemJson('i-2', name: 'Crown', price: 500),
  ];
  List<Map<String, Object?>> inventory = [];
  List<Map<String, Object?>> receivedGifts = [];

  /// Status of `GET /wallet` (e.g. 503).
  int walletStatus = 200;

  /// Status and error key of `POST /shop/purchase` (200 = success).
  int purchaseStatus = 200;
  String purchaseKey = 'res_err_insufficient_balance';

  /// Answered as `animationUrl`.
  String animationUrl = '';

  /// When set, `POST /shop/purchase` waits for it before answering.
  Completer<void>? holdPurchase;

  /// Bodies of `POST /shop/purchase`, in order.
  final List<Map<String, Object?>> purchaseBodies = [];

  ResponseBody _error(int status, String key) => jsonBody({
    'requestId': 'req-err',
    'success': false,
    'code': 1,
    'key': key,
    'message': 'Server says $key',
  }, status: status);

  String get _spark =>
      '${sparkCents ~/ 100}.${(sparkCents % 100).toString().padLeft(2, '0')}';

  ResponseBody _wallet() => okEnvelope({
    'account': {
      'id': 'a-1',
      'ownerId': 'u-1',
      'type': 'user_wallet',
      'status': 'active',
      'createdAt': '2026-11-04T00:00:00Z',
    },
    'balances': [
      {
        'accountId': 'a-1',
        'currencyCode': 'SPARK',
        'balance': _spark,
        'lastEntryId': null,
      },
      if (flareBalance != null)
        {
          'accountId': 'a-1',
          'currencyCode': 'FLARE',
          'balance': flareBalance,
          'lastEntryId': null,
        },
    ],
  });

  Future<ResponseBody> _purchase(RequestOptions o) async {
    final body = Map<String, Object?>.of(o.data as Map<String, Object?>);
    purchaseBodies.add(body);
    await holdPurchase?.future;
    if (purchaseStatus != 200) return _error(purchaseStatus, purchaseKey);
    final item = shopItems.firstWhere((i) => i['id'] == body['shopItemId']);
    sparkCents -= (item['price']! as int) * 100;
    if (body['recipientUserId'] == null) {
      inventory = [
        ...inventory,
        {
          'id': 'inv-${inventory.length}',
          'userId': 'u-1',
          'shopItemId': item['id'],
          'quantity': 1,
          'updatedAt': '2026-11-04T10:00:00Z',
        },
      ];
    }
    return okEnvelope({
      'giftId': body['recipientUserId'] == null ? null : 'g-1',
      'animationUrl': animationUrl,
    });
  }

  ResponseBody _page(List<Map<String, Object?>> all, RequestOptions o) {
    final page = o.queryParameters['page'] as int;
    final limit = o.queryParameters['limit'] as int;
    return okEnvelope(all.skip(page * limit).take(limit).toList());
  }

  @override
  Future<ResponseBody> handle(RequestOptions o) async {
    final path = o.path;
    if (o.method == 'GET') {
      switch (path) {
        case ApiEndpoints.wallet:
          return walletStatus == 200 ? _wallet() : _error(walletStatus, 'x');
        case ApiEndpoints.transactions:
          return _page(transactions, o);
        case ApiEndpoints.shopItems:
          return okEnvelope(shopItems);
        case ApiEndpoints.inventory:
          return _page(inventory, o);
      }
      if (path.startsWith('${ApiEndpoints.user}/') &&
          path.endsWith('/gifts/received')) {
        return _page(receivedGifts, o);
      }
    }
    if (o.method == 'POST' && path == ApiEndpoints.shopPurchase) {
      return _purchase(o);
    }
    return super.handle(o);
  }
}
