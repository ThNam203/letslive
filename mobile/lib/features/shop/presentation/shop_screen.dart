/// The shop (web `/shop`): the catalogue with a Buy button per item. Reached
/// from the Profile tab. A purchase is sent once; while it runs every Buy
/// button is disabled, and afterwards the balance and inventory reload (see
/// `PurchaseNotifier`). An error such as "Insufficient balance." is shown as
/// returned by the server's error key.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_item_tile.dart';
import 'package:letslive_mobile/features/shop/shop_providers.dart';

/// The shop screen.
class ShopScreen extends ConsumerWidget {
  const ShopScreen({super.key});

  Future<void> _buy(BuildContext context, WidgetRef ref, ShopItem item) async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final response = await ref.read(purchaseProvider.notifier).buy(item.id);
    if (response == null || !context.mounted) return;
    if (response.success) {
      messenger.showSnackBar(
        SnackBar(content: Text(l10n.shop_shop_purchase_success)),
      );
    } else {
      showActionError(context, response);
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final items = ref.watch(shopItemsProvider);
    final pendingId = ref.watch(purchaseProvider);
    final status = feedStatusBody(
      value: items,
      onRetry: () => ref.invalidate(shopItemsProvider),
      empty: EmptyNotice(
        icon: Icons.storefront_outlined,
        title: l10n.shop_shop_empty,
      ),
    );
    return Scaffold(
      appBar: AppBar(title: Text(l10n.shop_shop_page_title)),
      body: status != null
          ? ListView(children: [status])
          : ListView(
              padding: const EdgeInsets.symmetric(vertical: 8),
              children: [
                for (final item in items.requireValue)
                  ShopItemTile(
                    key: ValueKey('shop-item-${item.id}'),
                    name: item.name,
                    imageUrl: item.imageUrl,
                    description: item.description,
                    badge: l10n.shop_shop_price_label(item.price),
                    action: FilledButton(
                      key: ValueKey('buy-${item.id}'),
                      onPressed: pendingId == null
                          ? () => _buy(context, ref, item)
                          : null,
                      child: Text(
                        pendingId == item.id
                            ? l10n.shop_shop_gift_sending
                            : l10n.shop_shop_buy_button,
                      ),
                    ),
                  ),
              ],
            ),
    );
  }
}
