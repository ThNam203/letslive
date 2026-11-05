import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_entries_list.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_item_tile.dart';
import 'package:letslive_mobile/features/shop/shop_providers.dart';

/// The Inventory tab of the wallet (web `/wallet/inventory`): the items the
/// user owns and how many.
class InventoryTab extends ConsumerWidget {
  const InventoryTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return ShopEntriesList<InventoryItem>(
      feed: ref.watch(inventoryProvider),
      onRetry: () => ref.invalidate(inventoryProvider),
      onLoadMore: () => ref.read(inventoryProvider.notifier).loadMore(),
      emptyTitle: l10n.shop_inventory_empty,
      shopItemIdOf: (entry) => entry.shopItemId,
      tileOf: (entry, item) => ShopItemTile(
        key: ValueKey('inventory-${entry.id}'),
        name: item?.name ?? l10n.shop_shop_unknown_item,
        imageUrl: item?.imageUrl,
        badge: l10n.shop_inventory_quantity_label(entry.quantity),
      ),
    );
  }
}
