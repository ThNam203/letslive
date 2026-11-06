import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/wallet/data/wallet_models.dart';

/// Localized name of a transaction type, like web's
/// `t("wallet:transactions.type.<type>")`. A type this app version does not
/// know (the backend added one) shows as sent.
String transactionTypeLabel(AppLocalizations l10n, String type) =>
    switch (type) {
      'reward' => l10n.wallet_transactions_type_reward,
      'purchase' => l10n.wallet_transactions_type_purchase,
      'deposit' => l10n.wallet_transactions_type_deposit,
      'trade' => l10n.wallet_transactions_type_trade,
      'donate' => l10n.wallet_transactions_type_donate,
      'refund' => l10n.wallet_transactions_type_refund,
      'fee' => l10n.wallet_transactions_type_fee,
      'adjustment' => l10n.wallet_transactions_type_adjustment,
      _ => type,
    };

/// Localized name of a transaction status; unknown ones show as sent.
String _statusLabel(AppLocalizations l10n, String status) => switch (status) {
  'created' => l10n.wallet_transactions_status_created,
  'processing' => l10n.wallet_transactions_status_processing,
  'completed' => l10n.wallet_transactions_status_completed,
  'failed' => l10n.wallet_transactions_status_failed,
  'cancelled' => l10n.wallet_transactions_status_cancelled,
  _ => status,
};

IconData _typeIcon(String type) => switch (type) {
  'reward' => Icons.emoji_events_outlined,
  'purchase' => Icons.shopping_cart_outlined,
  'deposit' => Icons.credit_card,
  'trade' => Icons.swap_horiz,
  'donate' => Icons.favorite_border,
  'refund' => Icons.replay,
  'adjustment' => Icons.build_outlined,
  _ => Icons.attach_money,
};

/// One transaction (web `TransactionRow`): type, date, the signed net amount
/// (green for a credit, red for a debit) and the status.
class TransactionRow extends StatelessWidget {
  const TransactionRow({required this.transaction, super.key});

  /// The transaction shown.
  final WalletTransaction transaction;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final locale = Localizations.localeOf(context).toString();
    final net = transaction.netAmount;
    final color = net.isNegative ? Colors.red : Colors.green;
    // A plain row, not a ListTile: a tile's trailing column is capped to the
    // tile's height and overflows at large text sizes.
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(_typeIcon(transaction.type)),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  transactionTypeLabel(l10n, transaction.type),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.bodyLarge,
                ),
                Text(
                  DateFormat.yMMMd(locale)
                      .add_jm()
                      .format(transaction.createdAt.toLocal()),
                  style: theme.textTheme.bodySmall,
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 120),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  net.format(locale, withPlus: true),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.titleSmall?.copyWith(color: color),
                ),
                Text(
                  _statusLabel(l10n, transaction.status),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.labelSmall,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
