import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/wallet/presentation/transaction_row.dart';
import 'package:letslive_mobile/features/wallet/wallet_providers.dart';

/// The types offered as filters, in web's order
/// (`FILTER_OPTIONS` in `web/app/(main)/wallet/transactions/page.tsx`).
/// Deposits are not offered there either.
const List<String> _filterTypes = [
  'donate',
  'purchase',
  'reward',
  'refund',
  'trade',
  'fee',
];

/// The Transactions tab (web `/wallet/transactions`): all transactions,
/// paged as the list scrolls, with type filter chips. The filter only hides
/// loaded rows, as on web; the list keeps loading while the end is in view.
class WalletTransactionsTab extends ConsumerStatefulWidget {
  const WalletTransactionsTab({super.key});

  @override
  ConsumerState<WalletTransactionsTab> createState() =>
      _WalletTransactionsTabState();
}

class _WalletTransactionsTabState extends ConsumerState<WalletTransactionsTab> {
  /// The selected type; null shows all.
  String? _filter;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final feed = ref.watch(transactionsProvider);
    final loaded = feed.value;
    final status = feedStatusBody(
      value: feed.whenData((f) => f.items),
      onRetry: () => ref.invalidate(transactionsProvider),
      empty: EmptyNotice(
        icon: Icons.receipt_long_outlined,
        title: l10n.wallet_transactions_no_transactions,
        description: l10n.wallet_transactions_no_transactions_description,
      ),
    );
    final chips = Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 0),
      child: Wrap(
        spacing: 8,
        children: [
          ChoiceChip(
            key: const ValueKey('filter-all'),
            label: Text(l10n.wallet_transactions_filter_all),
            selected: _filter == null,
            onSelected: (_) => setState(() => _filter = null),
          ),
          for (final type in _filterTypes)
            ChoiceChip(
              key: ValueKey('filter-$type'),
              label: Text(transactionTypeLabel(l10n, type)),
              selected: _filter == type,
              onSelected: (_) => setState(() => _filter = type),
            ),
        ],
      ),
    );
    if (status != null || loaded == null) {
      return ListView(children: [chips, ?status]);
    }
    final shown = [
      for (final t in loaded.items)
        if (_filter == null || t.type == _filter) t,
    ];
    void loadMore() => ref.read(transactionsProvider.notifier).loadMore();
    if (shown.isEmpty) {
      // The filter hides every loaded row: say so, and let the user ask for
      // more instead of paging through the whole history by itself.
      return ListView(
        children: [
          chips,
          EmptyNotice(
            icon: Icons.receipt_long_outlined,
            title: l10n.wallet_transactions_no_transactions,
          ),
          if (loaded.hasMore)
            Center(
              child: TextButton(
                key: const ValueKey('load-more'),
                onPressed: loaded.isLoadingMore ? null : loadMore,
                child: Text(l10n.wallet_transactions_load_more),
              ),
            ),
        ],
      );
    }
    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(transactionsProvider),
      child: ListView.builder(
        itemCount: shown.length + 2,
        itemBuilder: (context, index) {
          if (index == 0) return chips;
          if (index == shown.length + 1) {
            return PagedFooter(feed: loaded, onLoadMore: loadMore);
          }
          return TransactionRow(transaction: shown[index - 1]);
        },
      ),
    );
  }
}
