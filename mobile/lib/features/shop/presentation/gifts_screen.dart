/// The gifts a user received (web `/users/[userId]/gifts`), opened from the
/// About tab of their profile.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_entries_list.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_item_tile.dart';
import 'package:letslive_mobile/features/shop/shop_providers.dart';

/// The received-gifts screen of [userId].
class GiftsScreen extends ConsumerWidget {
  const GiftsScreen({required this.userId, super.key});

  /// The recipient whose gifts are listed.
  final String userId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final provider = receivedGiftsProvider(userId);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.shop_gifts_received_page_title)),
      body: ShopEntriesList<Gift>(
        feed: ref.watch(provider),
        onRetry: () => ref.invalidate(provider),
        onLoadMore: () => ref.read(provider.notifier).loadMore(),
        emptyTitle: l10n.shop_gifts_received_empty,
        shopItemIdOf: (gift) => gift.shopItemId,
        tileOf: (gift, item) => ShopItemTile(
          key: ValueKey('gift-${gift.id}'),
          name: item?.name ?? l10n.shop_shop_unknown_item,
          imageUrl: item?.imageUrl,
          description: gift.message == null ? null : '"${gift.message}"',
          badge: l10n.shop_gifts_received_quantity_label(gift.quantity),
        ),
      ),
    );
  }
}
