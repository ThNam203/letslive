import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../l10n/app_localizations.dart';
import '../../../models/wallet.dart';
import '../../../providers.dart';
import 'wallet_transaction_list.dart';

class WalletTransactionsScreen extends ConsumerStatefulWidget {
  const WalletTransactionsScreen({super.key});

  @override
  ConsumerState<WalletTransactionsScreen> createState() =>
      _WalletTransactionsScreenState();
}

class _WalletTransactionsScreenState
    extends ConsumerState<WalletTransactionsScreen> {
  final List<Transaction> _transactions = [];
  int _page = 0;
  int _total = 0;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _fetch(0);
  }

  Future<void> _fetch(int page) async {
    setState(() => _isLoading = true);
    try {
      final repo = ref.read(walletRepositoryProvider);
      final res = await repo.getTransactions(page: page, pageSize: 20);
      if (res.success && res.data != null) {
        setState(() {
          if (page == 0) {
            _transactions
              ..clear()
              ..addAll(res.data!);
          } else {
            _transactions.addAll(res.data!);
          }
          _page = page;
          _total = res.meta?.total ?? 0;
        });
      }
    } catch (_) {
      // handled by UI
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final hasMore = _transactions.length < _total;

    return FScaffold(
      header: FHeader(title: Text(l10n.walletTransactionsTitle)),
      child: _isLoading && _transactions.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (_transactions.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 48),
                    child: Center(
                      child: Text(l10n.walletNoTransactions),
                    ),
                  )
                else ...[
                  WalletTransactionList(transactions: _transactions),
                  if (hasMore)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      child: FButton(
                        variant: FButtonVariant.outline,
                        onPress: _isLoading
                            ? null
                            : () => _fetch(_page + 1),
                        child: _isLoading
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : Text(l10n.walletLoadMore),
                      ),
                    ),
                ],
              ],
            ),
    );
  }
}
