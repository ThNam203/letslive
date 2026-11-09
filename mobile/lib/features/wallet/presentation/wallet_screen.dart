/// The wallet (web `/wallet`): tabs Overview, Transactions and Inventory,
/// opened from the Profile tab. Web's Deposit tab is not ported on purpose:
/// deposits and payments are out of scope for the mobile release, so nothing
/// here leads to adding funds.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/shop/presentation/inventory_tab.dart';
import 'package:letslive_mobile/features/wallet/presentation/overview_tab.dart';
import 'package:letslive_mobile/features/wallet/presentation/transactions_tab.dart';

/// The wallet screen.
class WalletScreen extends ConsumerWidget {
  const WalletScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return DefaultTabController(
      length: 3,
      child: Scaffold(
        appBar: AppBar(
          title: Text(l10n.wallet_page_title),
          bottom: TabBar(
            tabs: [
              Tab(text: l10n.wallet_navigation_overview),
              Tab(text: l10n.wallet_navigation_transactions),
              Tab(text: l10n.wallet_navigation_inventory),
            ],
          ),
        ),
        body: TabBarView(
          children: [
            Builder(
              builder: (context) => WalletOverviewTab(
                onViewAll: () => DefaultTabController.of(context).animateTo(1),
              ),
            ),
            const WalletTransactionsTab(),
            const InventoryTab(),
          ],
        ),
      ),
    );
  }
}
