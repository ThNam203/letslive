import 'package:flutter/material.dart';
import 'package:forui/forui.dart';

import '../../../l10n/app_localizations.dart';
import '../../../models/wallet.dart';

class WalletTransactionList extends StatelessWidget {
  final List<Transaction> transactions;

  const WalletTransactionList({super.key, required this.transactions});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: transactions
          .map((txn) => _TransactionTile(transaction: txn))
          .toList(),
    );
  }
}

class _TransactionTile extends StatelessWidget {
  final Transaction transaction;

  const _TransactionTile({required this.transaction});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final typography = context.theme.typography;
    final net = transaction.netAmount;
    final isPositive = net >= 0;

    final typeLabel = _transactionTypeLabel(l10n, transaction.type);
    final statusLabel = _transactionStatusLabel(l10n, transaction.status);
    final formattedAmount =
        '${isPositive ? '+' : ''}${net.toStringAsFixed(net.truncateToDouble() == net ? 0 : 2)}';

    final date = DateTime.tryParse(transaction.createdAt);
    final dateStr = date != null
        ? '${date.day}/${date.month}/${date.year} ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}'
        : '';

    return FTile(
      prefix: Icon(_typeIcon(transaction.type), size: 20),
      title: Text(typeLabel),
      subtitle: Text(dateStr, style: typography.xs),
      suffix: Column(
        crossAxisAlignment: CrossAxisAlignment.end,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            formattedAmount,
            style: typography.sm.copyWith(
              fontWeight: FontWeight.w600,
              color: isPositive ? Colors.green : Colors.red,
            ),
          ),
          Text(statusLabel, style: typography.xs),
        ],
      ),
    );
  }

  IconData _typeIcon(TransactionType type) => switch (type) {
    TransactionType.reward => Icons.emoji_events,
    TransactionType.purchase => Icons.shopping_cart,
    TransactionType.trade => Icons.swap_horiz,
    TransactionType.donate => Icons.favorite,
    TransactionType.refund => Icons.replay,
    TransactionType.fee => Icons.receipt,
    TransactionType.adjustment => Icons.tune,
  };

  String _transactionTypeLabel(AppLocalizations l10n, TransactionType type) =>
      switch (type) {
        TransactionType.reward => l10n.walletTxnReward,
        TransactionType.purchase => l10n.walletTxnPurchase,
        TransactionType.trade => l10n.walletTxnTrade,
        TransactionType.donate => l10n.walletTxnDonate,
        TransactionType.refund => l10n.walletTxnRefund,
        TransactionType.fee => l10n.walletTxnFee,
        TransactionType.adjustment => l10n.walletTxnAdjustment,
      };

  String _transactionStatusLabel(
    AppLocalizations l10n,
    TransactionStatus status,
  ) =>
      switch (status) {
        TransactionStatus.created => l10n.walletStatusCreated,
        TransactionStatus.processing => l10n.walletStatusProcessing,
        TransactionStatus.completed => l10n.walletStatusCompleted,
        TransactionStatus.failed => l10n.walletStatusFailed,
        TransactionStatus.cancelled => l10n.walletStatusCancelled,
      };
}
