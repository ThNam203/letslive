import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_item_tile.dart';
import 'package:letslive_mobile/features/shop/shop_providers.dart';

/// A paged list of things that point at a catalogue item (inventory rows,
/// received gifts). Shows a spinner, an error with retry, [emptyTitle], or
/// one [ShopItemTile] per entry from [tileOf], which receives the entry and
/// its catalogue item (found by [shopItemIdOf]) (null when it left the catalogue or the catalogue
/// failed to load: web then shows "Unknown item" too).
class ShopEntriesList<T> extends ConsumerWidget {
  const ShopEntriesList({
    required this.feed,
    required this.onRetry,
    required this.onLoadMore,
    required this.emptyTitle,
    required this.shopItemIdOf,
    required this.tileOf,
    super.key,
  });

  /// The entries loaded so far.
  final AsyncValue<PagedFeed<T>> feed;

  /// Reloads the first page.
  final VoidCallback onRetry;

  /// Loads the next page; also the retry after a failed one.
  final VoidCallback onLoadMore;

  /// Text of the empty state.
  final String emptyTitle;

  /// The catalogue id an entry points at.
  final String Function(T entry) shopItemIdOf;

  /// Builds the row of one entry.
  final ShopItemTile Function(T entry, ShopItem? item) tileOf;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final loaded = feed.value;
    if (loaded == null || (feed.hasError && !feed.isLoading)) {
      return ListView(
        children: [
          feedStatusBody<T>(
                value: feed.whenData((f) => f.items),
                onRetry: onRetry,
              ) ??
              const SizedBox.shrink(),
        ],
      );
    }
    if (loaded.items.isEmpty) {
      return ListView(
        children: [
          EmptyNotice(icon: Icons.inventory_2_outlined, title: emptyTitle),
        ],
      );
    }
    final catalogue = {
      for (final item in ref.watch(shopItemsProvider).value ?? <ShopItem>[])
        item.id: item,
    };
    return ListView.builder(
      padding: const EdgeInsets.symmetric(vertical: 8),
      itemCount: loaded.items.length + 1,
      itemBuilder: (context, index) {
        if (index == loaded.items.length) {
          return PagedFooter<T>(feed: loaded, onLoadMore: onLoadMore);
        }
        final entry = loaded.items[index];
        return tileOf(entry, catalogue[shopItemIdOf(entry)]);
      },
    );
  }
}
