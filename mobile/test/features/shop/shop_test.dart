import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/shop/presentation/gift_sheet.dart';
import 'package:letslive_mobile/features/shop/shop_providers.dart';
import 'package:letslive_mobile/features/wallet/wallet_providers.dart';

import '../auth/auth_test_app.dart';
import '../profile/profile_test_app.dart';
import '../wallet/wallet_test_app.dart';

final _buyRose = find.byKey(const ValueKey('buy-i-1'));
final _giftButton = find.byKey(const ValueKey('gift-button'));
final _giftRose = find.byKey(const ValueKey('gift-item-i-1'));

void main() {
  group('shop', () {
    testWidgets('lists the catalogue with prices and sends one purchase', (
      tester,
    ) async {
      final server = FakeWalletServer();
      await pumpRoute(tester, server, AppRoutes.shop);
      expect(find.text('Rose'), findsOneWidget);
      expect(find.text('A red rose'), findsOneWidget);
      expect(find.text('10 SPARK'), findsOneWidget);
      expect(find.text('500 SPARK'), findsOneWidget);

      await tester.tap(_buyRose);
      await tester.pumpAndSettle();

      expect(server.purchaseBodies, [
        {'shopItemId': 'i-1', 'quantity': 1},
      ]);
      expect(find.text('Purchase successful!'), findsOneWidget);
    });

    testWidgets('insufficient balance shows web\'s error key message', (
      tester,
    ) async {
      final server = FakeWalletServer()..purchaseStatus = 400;
      await pumpRoute(tester, server, AppRoutes.shop);

      await tester.tap(_buyRose);
      await tester.pumpAndSettle();

      expect(find.text('Insufficient balance.'), findsOneWidget);
      expect(find.text('Purchase successful!'), findsNothing);
      // The text is the translation of the key, with no way to add funds.
      expect(find.textContaining('deposit', findRichText: true), findsNothing);
    });

    testWidgets('an unknown error key falls back to the server message', (
      tester,
    ) async {
      final server = FakeWalletServer()
        ..purchaseStatus = 500
        ..purchaseKey = 'res_err_user_service_error';
      await pumpRoute(tester, server, AppRoutes.shop);

      await tester.tap(_buyRose);
      await tester.pumpAndSettle();

      expect(
        find.text('Could not deliver the item, your payment was refunded.'),
        findsOneWidget,
      );
    });

    testWidgets('a second tap while buying is ignored and buttons are '
        'disabled', (tester) async {
      final server = FakeWalletServer()..holdPurchase = Completer<void>();
      await pumpRoute(tester, server, AppRoutes.shop);

      await tester.tap(_buyRose);
      await tester.pump();
      expect(find.text('Sending...'), findsOneWidget);
      await tester.tap(_buyRose, warnIfMissed: false);
      await tester.tap(
        find.byKey(const ValueKey('buy-i-2')),
        warnIfMissed: false,
      );
      await tester.pump();

      server.holdPurchase!.complete();
      await tester.pumpAndSettle();

      expect(server.purchaseBodies, hasLength(1));
      expect(find.text('Sending...'), findsNothing);
    });

    testWidgets('a failed purchase is not retried by itself', (tester) async {
      final server = FakeWalletServer()..purchaseStatus = 503;
      final app = await pumpRoute(tester, server, AppRoutes.shop);

      await tester.tap(_buyRose);
      await tester.pumpAndSettle();
      await tester.pump(const Duration(seconds: 30));

      expect(app.requestsTo(ApiEndpoints.shopPurchase), hasLength(1));
    });

    testWidgets('purchase refreshes balance and inventory', (tester) async {
      final server = FakeWalletServer()..sparkCents = 10000;
      final app = await pumpRoute(tester, server, AppRoutes.shop);
      // Keep what the wallet screens would show alive and loaded.
      final container = app.container;
      final balance = container.listen(walletOverviewProvider, (_, _) {});
      final inventory = container.listen(inventoryProvider, (_, _) {});
      addTearDown(balance.close);
      addTearDown(inventory.close);
      await tester.pumpAndSettle();
      expect(
        balance.read().requireValue.balanceOf('SPARK').format('en'),
        '100',
      );
      expect(inventory.read().requireValue.items, isEmpty);

      await tester.tap(_buyRose);
      await tester.pumpAndSettle();

      expect(balance.read().requireValue.balanceOf('SPARK').format('en'), '90');
      expect(inventory.read().requireValue.items, hasLength(1));
      expect(app.requestsTo(ApiEndpoints.wallet), hasLength(2));
      expect(app.requestsTo(ApiEndpoints.inventory), hasLength(2));
    });

    testWidgets('a failed purchase still refreshes the balance', (
      tester,
    ) async {
      final server = FakeWalletServer()..purchaseStatus = 400;
      final app = await pumpRoute(tester, server, AppRoutes.shop);
      final balance = app.container.listen(walletOverviewProvider, (_, _) {});
      addTearDown(balance.close);
      await tester.pumpAndSettle();

      await tester.tap(_buyRose);
      await tester.pumpAndSettle();

      expect(app.requestsTo(ApiEndpoints.wallet), hasLength(2));
    });

    testWidgets('an empty shop and a failed load', (tester) async {
      final server = FakeWalletServer()..shopItems = [];
      await pumpRoute(tester, server, AppRoutes.shop);
      expect(find.text('No items available right now.'), findsOneWidget);
    });

    testWidgets('long names do not overflow at 320px with text scale 2', (
      tester,
    ) async {
      final long = 'Một món quà rất dài không dấu 🎁 ${'x' * 200}';
      final server = FakeWalletServer()
        ..shopItems = [shopItemJson('i-1', name: long, description: long)];
      await pumpRoute(
        tester,
        server,
        AppRoutes.shop,
        width: 320,
        height: 568,
        textScale: 2,
      );
      expect(tester.takeException(), isNull);
      await tester.tap(find.byKey(const ValueKey('buy-i-1')));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  });

  group('gift sheet', () {
    testWidgets('sends the picked item to the channel owner', (tester) async {
      final server = FakeWalletServer();
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));

      await tester.tap(_giftButton);
      await tester.pumpAndSettle();
      expect(find.byType(GiftSheet), findsOneWidget);
      expect(find.textContaining('Pick an item to send'), findsOneWidget);
      expect(find.textContaining('other'), findsWidgets);

      await tester.tap(_giftRose);
      await tester.pumpAndSettle();

      expect(server.purchaseBodies, [
        {'shopItemId': 'i-1', 'quantity': 1, 'recipientUserId': 'u-2'},
      ]);
      expect(find.text('Gift sent!'), findsOneWidget);
      // A gift does not go into the sender's inventory.
      expect(server.inventory, isEmpty);
    });

    testWidgets('insufficient balance is shown in the sheet', (tester) async {
      final server = FakeWalletServer()..purchaseStatus = 400;
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await tester.tap(_giftButton);
      await tester.pumpAndSettle();

      await tester.tap(_giftRose);
      await tester.pumpAndSettle();

      expect(find.text('Insufficient balance.'), findsOneWidget);
      expect(find.text('Gift sent!'), findsNothing);
      expect(find.byType(GiftSheet), findsOneWidget);
    });

    testWidgets('a second tap while sending is ignored', (tester) async {
      final server = FakeWalletServer()..holdPurchase = Completer<void>();
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await tester.tap(_giftButton);
      await tester.pumpAndSettle();

      await tester.tap(_giftRose);
      await tester.pump();
      expect(find.text('Sending...'), findsOneWidget);
      await tester.tap(_giftRose, warnIfMissed: false);
      await tester.tap(
        find.byKey(const ValueKey('gift-item-i-2')),
        warnIfMissed: false,
      );
      server.holdPurchase!.complete();
      await tester.pumpAndSettle();

      expect(server.purchaseBodies, hasLength(1));
    });

    testWidgets('the animation shows, then the sheet closes after 3 s', (
      tester,
    ) async {
      final server = FakeWalletServer()
        ..animationUrl = 'http://localhost:8000/files/shop/i-1.gif';
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await tester.tap(_giftButton);
      await tester.pumpAndSettle();

      await tester.tap(_giftRose);
      for (var i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 100));
      }
      expect(find.byKey(const ValueKey('gift-animation')), findsOneWidget);

      await tester.pump(const Duration(seconds: 3));
      await tester.pumpAndSettle();
      expect(find.byType(GiftSheet), findsNothing);
    });

    testWidgets('your own channel has no gift button', (tester) async {
      await pumpRoute(tester, FakeWalletServer(), AppRoutes.channel('u-1'));
      expect(_giftButton, findsNothing);
    });

    testWidgets('an empty catalogue says so', (tester) async {
      final server = FakeWalletServer()..shopItems = [];
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await tester.tap(_giftButton);
      await tester.pumpAndSettle();
      expect(find.text('No items in your inventory.'), findsOneWidget);
    });

    testWidgets('a long recipient name does not overflow at 320px with text '
        'scale 2', (tester) async {
      final server = FakeWalletServer();
      server.users['u-2'] = {
        ...server.users['u-2']!,
        'username': 'Người dùng tên rất dài không dấu 🎉 ${'y' * 100}',
      };
      await pumpRoute(
        tester,
        server,
        AppRoutes.channel('u-2'),
        width: 320,
        height: 568,
        textScale: 2,
      );
      await tester.tap(_giftButton);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  });

  group('received gifts', () {
    Map<String, Object?> gift(String id, String item, int qty, {String? msg}) =>
        {
          'id': id,
          'senderUserId': 'u-9',
          'recipientUserId': 'u-2',
          'shopItemId': item,
          'quantity': qty,
          'message': msg,
          'sentAt': '2026-11-04T10:00:00Z',
        };

    testWidgets('the About tab opens them', (tester) async {
      final server = FakeWalletServer()
        ..receivedGifts = [
          gift('g-1', 'i-1', 2, msg: 'Great stream'),
          gift('g-2', 'gone', 1),
        ];
      await pumpRoute(tester, server, AppRoutes.channel('u-2'));
      await tester.tap(find.text('About').last);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('open-gifts')));
      await tester.pumpAndSettle();

      expect(find.text('Gift Collection'), findsWidgets);
      expect(find.text('Rose'), findsOneWidget);
      expect(find.text('×2'), findsOneWidget);
      expect(find.text('"Great stream"'), findsOneWidget);
      expect(find.text('Unknown item'), findsOneWidget);
    });

    testWidgets('an empty list says so', (tester) async {
      await pumpRoute(tester, FakeWalletServer(), AppRoutes.gifts('u-2'));
      expect(find.text('No gifts received yet.'), findsOneWidget);
    });

    testWidgets('pages load as the list scrolls', (tester) async {
      final server = FakeWalletServer()
        ..receivedGifts = [
          for (var i = 0; i < 25; i++) gift('g-$i', 'i-1', i + 1),
        ];
      final app = await pumpRoute(
        tester,
        server,
        AppRoutes.gifts('u-2'),
        height: 800,
      );
      await tester.scrollUntilVisible(find.text('×25'), 300);
      await tester.pumpAndSettle();

      final pages = [
        for (final r in app.adapter.requests)
          if (r.path.endsWith('/gifts/received')) r.queryParameters['page'],
      ];
      expect(pages, [0, 1]);
    });
  });

  testWidgets('the next account never sees the previous account\'s '
      'inventory', (tester) async {
    final server = FakeWalletServer()
      ..inventory = [
        {
          'id': 'inv-a',
          'userId': 'u-1',
          'shopItemId': 'i-2',
          'quantity': 4,
          'updatedAt': '2026-11-04T00:00:00Z',
        },
      ];
    final app = await pumpRoute(tester, server, AppRoutes.wallet);
    await tester.tap(find.text('Inventory'));
    await tester.pumpAndSettle();
    expect(find.text('Crown'), findsOneWidget);

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);

    server.inventory = [];
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;
    unawaited(app.container.read(appRouterProvider).push(AppRoutes.wallet));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Inventory'));
    await tester.pumpAndSettle();

    expect(app.requestsTo(ApiEndpoints.inventory), hasLength(2));
    expect(find.text('Crown'), findsNothing);
    expect(find.text("You don't own any items yet."), findsOneWidget);
  });
}
