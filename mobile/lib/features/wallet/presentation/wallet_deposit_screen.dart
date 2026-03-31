import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../l10n/app_localizations.dart';
import '../../../models/wallet.dart';
import '../../../providers.dart';
import '../../../core/utils/api_error_localizer.dart';

class WalletDepositScreen extends ConsumerStatefulWidget {
  const WalletDepositScreen({super.key});

  @override
  ConsumerState<WalletDepositScreen> createState() =>
      _WalletDepositScreenState();
}

class _WalletDepositScreenState extends ConsumerState<WalletDepositScreen> {
  CurrencyCode _currency = CurrencyCode.spark;
  PaymentProvider _provider = PaymentProvider.stripe;
  final _amountController = TextEditingController();
  bool _isSubmitting = false;
  String? _error;

  @override
  void dispose() {
    _amountController.dispose();
    super.dispose();
  }

  Future<void> _handleDeposit() async {
    final l10n = AppLocalizations.of(context);
    final amount = _amountController.text.trim();
    final num = double.tryParse(amount);

    if (num == null || num <= 0) {
      setState(() => _error = l10n.walletDepositErrorInvalidAmount);
      return;
    }
    if (num < 1) {
      setState(() => _error = l10n.walletDepositErrorMinAmount);
      return;
    }

    setState(() {
      _error = null;
      _isSubmitting = true;
    });

    try {
      final repo = ref.read(walletRepositoryProvider);
      final res = await repo.createDeposit(
        provider: _provider.name,
        currencyCode: _currency.value,
        amount: amount,
      );

      if (res.success && res.data != null) {
        final url = Uri.parse(res.data!.checkoutUrl);
        if (await canLaunchUrl(url)) {
          await launchUrl(url, mode: LaunchMode.externalApplication);
        }
        if (mounted) Navigator.pop(context);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(getLocalizedApiMessage(context, res.key)),
            ),
          );
        }
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.apiDefaultError)),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FScaffold(
      header: FHeader(title: Text(l10n.walletDepositTitle)),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Currency selection
          Text(
            l10n.walletDepositSelectCurrency,
            style: typography.sm.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: _CurrencyOption(
                  icon: Icons.bolt,
                  name: l10n.walletCurrencySpark,
                  isSelected: _currency == CurrencyCode.spark,
                  onTap: () =>
                      setState(() => _currency = CurrencyCode.spark),
                  colors: colors,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _CurrencyOption(
                  icon: Icons.diamond,
                  name: l10n.walletCurrencyFlare,
                  isSelected: _currency == CurrencyCode.flare,
                  onTap: () =>
                      setState(() => _currency = CurrencyCode.flare),
                  colors: colors,
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),

          // Amount
          Text(
            l10n.walletDepositAmount,
            style: typography.sm.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          FTextField(
            control: FTextFieldControl.managed(
              controller: _amountController,
              onChange: (_) {
                if (_error != null) setState(() => _error = null);
              },
            ),
            hint: l10n.walletDepositAmountPlaceholder,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
          ),
          if (_error != null)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: Text(
                _error!,
                style: typography.xs.copyWith(color: colors.destructive),
              ),
            ),
          const SizedBox(height: 24),

          // Provider selection
          Text(
            l10n.walletDepositSelectProvider,
            style: typography.sm.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          FTile(
            prefix: const Icon(Icons.credit_card),
            title: Text(l10n.walletDepositProviderStripe),
            suffix: _provider == PaymentProvider.stripe
                ? Icon(FIcons.check, color: colors.primary)
                : null,
            onPress: () =>
                setState(() => _provider = PaymentProvider.stripe),
          ),
          FTile(
            prefix: const Icon(Icons.account_balance_wallet),
            title: Text(l10n.walletDepositProviderPaypal),
            suffix: _provider == PaymentProvider.paypal
                ? Icon(FIcons.check, color: colors.primary)
                : null,
            onPress: () =>
                setState(() => _provider = PaymentProvider.paypal),
          ),
          const SizedBox(height: 32),

          // Submit
          FButton(
            onPress: _isSubmitting ? null : _handleDeposit,
            child: _isSubmitting
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : Text(l10n.walletDepositConfirm),
          ),
        ],
      ),
    );
  }
}

class _CurrencyOption extends StatelessWidget {
  final IconData icon;
  final String name;
  final bool isSelected;
  final VoidCallback onTap;
  final FColors colors;

  const _CurrencyOption({
    required this.icon,
    required this.name,
    required this.isSelected,
    required this.onTap,
    required this.colors,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? colors.primary : colors.border,
            width: isSelected ? 2 : 1,
          ),
        ),
        child: Column(
          children: [
            Icon(icon, color: isSelected ? colors.primary : colors.foreground),
            const SizedBox(height: 4),
            Text(
              name,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w500,
                color: isSelected ? colors.primary : colors.foreground,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
