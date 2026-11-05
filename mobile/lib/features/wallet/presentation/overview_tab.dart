import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_models.dart';
import 'package:letslive_mobile/features/wallet/domain/decimal_amount.dart';
import 'package:letslive_mobile/features/wallet/presentation/transaction_row.dart';
import 'package:letslive_mobile/features/wallet/wallet_providers.dart';

/// How many recent transactions the overview lists (web: 5).
const int _recentCount = 5;

/// The Overview tab (web `/wallet/overview`): a card per currency, then the
/// latest transactions. It only reads: there is no way to add funds here.
class WalletOverviewTab extends ConsumerWidget {
  const WalletOverviewTab({required this.onViewAll, super.key});

  /// Switches to the Transactions tab.
  final VoidCallback onViewAll;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final wallet = ref.watch(walletOverviewProvider);
    final transactions = ref.watch(transactionsProvider);
    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(walletOverviewProvider);
        ref.invalidate(transactionsProvider);
      },
      child: ListView(
        children: [
          SectionHeader(l10n.wallet_overview_title),
          if (wallet.hasError && !wallet.isLoading)
            FeedError(
              error: wallet.error!,
              onRetry: () => ref.invalidate(walletOverviewProvider),
            )
          else if (wallet.value == null)
            const Padding(
              padding: EdgeInsets.all(24),
              child: Center(child: CircularProgressIndicator()),
            )
          else ...[
            _BalanceCard(
              key: const ValueKey('balance-SPARK'),
              name: l10n.wallet_currency_spark,
              description: l10n.wallet_currency_spark_description,
              balance: wallet.requireValue.balanceOf('SPARK'),
              colors: const [Color(0xFFF59E0B), Color(0xFFEA580C)],
            ),
            _BalanceCard(
              key: const ValueKey('balance-FLARE'),
              name: l10n.wallet_currency_flare,
              description: l10n.wallet_currency_flare_description,
              balance: wallet.requireValue.balanceOf('FLARE'),
              colors: const [Color(0xFF8B5CF6), Color(0xFF4F46E5)],
            ),
          ],
          SectionHeader(l10n.wallet_overview_recent_transactions),
          ..._recent(context, ref, l10n, transactions),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton(
              onPressed: onViewAll,
              child: Text(l10n.wallet_overview_view_all),
            ),
          ),
        ],
      ),
    );
  }

  List<Widget> _recent(
    BuildContext context,
    WidgetRef ref,
    AppLocalizations l10n,
    AsyncValue<PagedFeed<WalletTransaction>> transactions,
  ) {
    final status = feedStatusBody(
      value: transactions.whenData((feed) => feed.items),
      onRetry: () => ref.invalidate(transactionsProvider),
      empty: EmptyNotice(
        icon: Icons.receipt_long_outlined,
        title: l10n.wallet_overview_no_transactions,
      ),
    );
    if (status != null) return [status];
    return [
      for (final transaction in transactions.requireValue.items.take(
        _recentCount,
      ))
        TransactionRow(transaction: transaction),
    ];
  }
}

/// Gradient card with a currency's name, balance and description (web
/// `BalanceCard`). The amount scales down instead of overflowing.
class _BalanceCard extends StatelessWidget {
  const _BalanceCard({
    required this.name,
    required this.description,
    required this.balance,
    required this.colors,
    super.key,
  });

  final String name;
  final String description;
  final DecimalAmount balance;
  final List<Color> colors;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final locale = Localizations.localeOf(context).toString();
    const onGradient = TextStyle(color: Colors.white);
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      child: DecoratedBox(
        decoration: BoxDecoration(
          gradient: LinearGradient(colors: colors),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(name, style: theme.textTheme.titleSmall?.merge(onGradient)),
              FittedBox(
                fit: BoxFit.scaleDown,
                alignment: Alignment.centerLeft,
                child: Text(
                  balance.format(locale),
                  style: theme.textTheme.headlineMedium?.merge(onGradient),
                ),
              ),
              Text(
                description,
                style: theme.textTheme.bodySmall?.merge(onGradient),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
