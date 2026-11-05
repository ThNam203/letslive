import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/router/app_router.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/presentation/login_screen.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_models.dart';

import '../auth/auth_test_app.dart';
import '../profile/profile_test_app.dart';
import 'wallet_test_app.dart';

Finder _balance(String currency, String text) => find.descendant(
  of: find.byKey(ValueKey('balance-$currency')),
  matching: find.text(text),
);

Future<void> _openTab(WidgetTester tester, String label) async {
  await tester.tap(
    find.descendant(of: find.byType(TabBar), matching: find.text(label)),
  );
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('overview shows both balances formatted like web, 0 for a '
      'missing currency, and the recent transactions', (tester) async {
    final server = FakeWalletServer()
      ..sparkCents = 123456
      ..transactions = [
        for (var i = 0; i < 7; i++)
          transactionJson('t-$i', type: 'donate', amount: '-2.50'),
      ];
    await pumpRoute(tester, server, AppRoutes.wallet);

    expect(_balance('SPARK', '1,234.56'), findsOneWidget);
    expect(_balance('FLARE', '0'), findsOneWidget);
    expect(find.text('Your Balances'), findsOneWidget);
    // Web lists five recent ones.
    expect(find.text('Donation'), findsNWidgets(5));
    expect(find.text('-2.5'), findsNWidgets(5));
  });

  testWidgets('a failed balance shows the error with a retry', (tester) async {
    final server = FakeWalletServer()..walletStatus = 503;
    await pumpRoute(tester, server, AppRoutes.wallet);
    expect(find.text('Try again'), findsOneWidget);

    server.walletStatus = 200;
    await tester.tap(find.text('Try again'));
    await tester.pumpAndSettle();

    expect(_balance('SPARK', '100'), findsOneWidget);
  });

  testWidgets('no deposit entry point exists in wallet overview', (
    tester,
  ) async {
    final server = FakeWalletServer()
      ..transactions = [transactionJson('t-1', type: 'donate')];
    final app = await pumpRoute(tester, server, AppRoutes.wallet);

    // Web has Overview / Transactions / Deposit / Inventory; here no Deposit.
    expect(
      find.descendant(of: find.byType(TabBar), matching: find.byType(Tab)),
      findsNWidgets(3),
    );
    // Nowhere on any tab: no text, button, link or tooltip about adding funds.
    for (final tab in ['Overview', 'Transactions', 'Inventory']) {
      await _openTab(tester, tab);
      expect(
        find.textContaining(RegExp('deposit', caseSensitive: false)),
        findsNothing,
        reason: tab,
      );
      expect(find.byType(FilledButton), findsNothing, reason: tab);
      expect(find.byType(ElevatedButton), findsNothing, reason: tab);
      expect(find.textContaining('web'), findsNothing, reason: tab);
    }
    // The app never calls the excluded endpoints.
    final paths = app.adapter.requests.map((r) => r.path);
    expect(paths.where((p) => p.startsWith('/deposits')), isEmpty);
    expect(paths.where((p) => p.startsWith('/payments')), isEmpty);
  });

  testWidgets('transactions paginate', (tester) async {
    final server = FakeWalletServer()
      ..transactions = [
        for (var i = 0; i < 25; i++)
          transactionJson('t-$i', amount: i == 24 ? '77.00' : '-1.00'),
      ];
    final app = await pumpRoute(tester, server, AppRoutes.wallet, height: 800);
    await _openTab(tester, 'Transactions');
    expect(find.text('+77'), findsNothing);

    await tester.scrollUntilVisible(
      find.text('+77'),
      300,
      scrollable: find.byType(Scrollable).last,
    );
    await tester.pumpAndSettle();

    final pages = [
      for (final r in app.requestsTo(ApiEndpoints.transactions))
        r.queryParameters['page'],
    ];
    expect(pages, [0, 1]);
    expect(find.text('+77'), findsOneWidget);
  });

  testWidgets('the type filter hides other loaded transactions', (
    tester,
  ) async {
    final server = FakeWalletServer()
      ..transactions = [
        transactionJson('t-1', type: 'donate'),
        transactionJson('t-2', type: 'reward', amount: '5.00'),
      ];
    await pumpRoute(tester, server, AppRoutes.wallet);
    await _openTab(tester, 'Transactions');
    expect(find.text('-10'), findsOneWidget);
    expect(find.text('+5'), findsOneWidget);

    await tester.tap(find.byKey(const ValueKey('filter-reward')));
    await tester.pumpAndSettle();
    expect(find.text('-10'), findsNothing);
    expect(find.text('+5'), findsOneWidget);
  });

  testWidgets('inventory lists owned items and tolerates unknown ones', (
    tester,
  ) async {
    final server = FakeWalletServer()
      ..inventory =
          [
                {'id': 'inv-1', 'shopItemId': 'i-1', 'quantity': 3},
                {'id': 'inv-2', 'shopItemId': 'gone', 'quantity': 1},
              ]
              .map(
                (e) => {
                  ...e,
                  'userId': 'u-1',
                  'updatedAt': '2026-11-04T00:00:00Z',
                },
              )
              .toList();
    await pumpRoute(tester, server, AppRoutes.wallet);
    await _openTab(tester, 'Inventory');

    expect(find.text('Rose'), findsOneWidget);
    expect(find.text('×3'), findsOneWidget);
    expect(find.text('Unknown item'), findsOneWidget);
  });

  testWidgets('an empty inventory says so', (tester) async {
    await pumpRoute(tester, FakeWalletServer(), AppRoutes.wallet);
    await _openTab(tester, 'Inventory');
    expect(find.text("You don't own any items yet."), findsOneWidget);
  });

  testWidgets('the Profile tab opens the wallet and the shop', (tester) async {
    await pumpRoute(tester, FakeWalletServer(), AppRoutes.profile);
    await tester.tap(find.byKey(const ValueKey('open-wallet')));
    await tester.pumpAndSettle();
    expect(find.text('Your Balances'), findsOneWidget);

    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('open-shop')));
    await tester.pumpAndSettle();
    expect(find.text('Crown'), findsOneWidget);
  });

  testWidgets('the next account never sees the previous account\'s '
      'wallet', (tester) async {
    final server = FakeWalletServer()
      ..sparkCents = 11100
      ..transactions = [transactionJson('t-a', type: 'reward', amount: '1')];
    final app = await pumpRoute(tester, server, AppRoutes.wallet);
    expect(_balance('SPARK', '111'), findsOneWidget);
    expect(find.text('Reward'), findsOneWidget);

    final loggedOut = app.container.read(sessionProvider.notifier).logout();
    await tester.pumpAndSettle();
    await loggedOut;
    expect(find.byType(LoginScreen), findsOneWidget);

    // Account B: another balance and no transactions.
    server
      ..sparkCents = 22200
      ..transactions = [];
    final loggedIn = app.container
        .read(sessionProvider.notifier)
        .login('b@example.com', testPassword);
    await tester.pumpAndSettle();
    await loggedIn;
    unawaited(app.container.read(appRouterProvider).push(AppRoutes.wallet));
    await tester.pumpAndSettle();

    expect(app.requestsTo(ApiEndpoints.wallet), hasLength(2));
    expect(_balance('SPARK', '222'), findsOneWidget);
    expect(find.text('Reward'), findsNothing);
  });

  testWidgets('long values do not overflow at 320px with text scale 2', (
    tester,
  ) async {
    final server = FakeWalletServer()
      ..sparkCents = 99999999999999
      ..flareBalance = '123456789012345.67'
      ..transactions = [
        transactionJson('t-1', amount: '-123456789012345.67'),
        transactionJson('t-2', type: 'brand_new_type_with_a_very_long_name'),
      ];
    await pumpRoute(
      tester,
      server,
      AppRoutes.wallet,
      width: 320,
      height: 568,
      textScale: 2,
    );
    expect(tester.takeException(), isNull);
    await _openTab(tester, 'Transactions');
    expect(tester.takeException(), isNull);
    await _openTab(tester, 'Inventory');
    expect(tester.takeException(), isNull);
  });

  test('wallet models parse strictly and sum entries exactly', () {
    expect(
      () => WalletOverview.fromJson({
        'balances': [
          {'currencyCode': 'SPARK', 'balance': '1e3'},
        ],
      }),
      throwsFormatException,
    );
    final withoutEntries = WalletTransaction.fromJson({
      ...transactionJson('t-1'),
      'entries': null,
    });
    expect(withoutEntries.entries, isEmpty);
    expect(withoutEntries.netAmount.format('en'), '0');

    final tx = WalletTransaction.fromJson({
      ...transactionJson('t-2'),
      'entries': [
        {'currencyCode': 'SPARK', 'amount': '0.10'},
        {'currencyCode': 'SPARK', 'amount': '0.20'},
      ],
    });
    expect(tx.netAmount.format('en'), '0.3');
  });
}
