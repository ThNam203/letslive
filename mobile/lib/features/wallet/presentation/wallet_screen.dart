import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/router/app_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/wallet.dart';
import '../../../providers.dart';
import 'wallet_transaction_list.dart';

class WalletScreen extends ConsumerStatefulWidget {
  const WalletScreen({super.key});

  @override
  ConsumerState<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends ConsumerState<WalletScreen> {
  WalletOverview? _wallet;
  List<Transaction> _recentTxns = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    setState(() => _isLoading = true);
    try {
      final repo = ref.read(walletRepositoryProvider);
      final results = await Future.wait([
        repo.getMyWallet(),
        repo.getTransactions(page: 0, pageSize: 5),
      ]);

      final walletRes = results[0];
      final txnRes = results[1];

      if (walletRes.success && walletRes.data != null) {
        _wallet = walletRes.data as WalletOverview;
      }
      if (txnRes.success && txnRes.data != null) {
        _recentTxns = txnRes.data as List<Transaction>;
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
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return FScaffold(
      header: FHeader(
        title: Text(l10n.walletTitle),
        suffixes: [
          FButton(
            onPress: () => context.push(AppRoutes.walletDeposit),
            prefix: const Icon(FIcons.plus),
            child: Text(l10n.walletDeposit),
          ),
        ],
      ),
      child: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _fetchData,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  // Balance cards
                  Row(
                    children: [
                      Expanded(
                        child: _BalanceCard(
                          icon: Icons.bolt,
                          name: l10n.walletCurrencySpark,
                          balance: _wallet?.balanceFor(CurrencyCode.spark) ?? '0',
                          gradientColors: [
                            Colors.amber.shade500,
                            Colors.orange.shade600,
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _BalanceCard(
                          icon: Icons.diamond,
                          name: l10n.walletCurrencyFlare,
                          balance: _wallet?.balanceFor(CurrencyCode.flare) ?? '0',
                          gradientColors: [
                            Colors.purple.shade500,
                            Colors.indigo.shade600,
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),

                  // Recent transactions
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        l10n.walletRecentTransactions,
                        style: typography.lg.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      FButton.raw(
                        onPress: () =>
                            context.push(AppRoutes.walletTransactions),
                        child: Text(
                          l10n.walletViewAll,
                          style: typography.sm.copyWith(
                            color: colors.primary,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),

                  if (_recentTxns.isEmpty)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 32),
                      child: Center(
                        child: Text(
                          l10n.walletNoTransactions,
                          style: typography.sm.copyWith(
                            color: isDark
                                ? AppColors.darkForegroundMuted
                                : AppColors.lightForegroundMuted,
                          ),
                        ),
                      ),
                    )
                  else
                    WalletTransactionList(transactions: _recentTxns),
                ],
              ),
            ),
    );
  }
}

class _BalanceCard extends StatelessWidget {
  final IconData icon;
  final String name;
  final String balance;
  final List<Color> gradientColors;

  const _BalanceCard({
    required this.icon,
    required this.name,
    required this.balance,
    required this.gradientColors,
  });

  @override
  Widget build(BuildContext context) {
    final num = double.tryParse(balance) ?? 0;
    final formatted = num.toStringAsFixed(num.truncateToDouble() == num ? 0 : 2);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: gradientColors,
        ),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: Colors.white.withValues(alpha: 0.6), size: 24),
          const SizedBox(height: 8),
          Text(
            name,
            style: const TextStyle(
              color: Colors.white70,
              fontSize: 12,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            formatted,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }
}
