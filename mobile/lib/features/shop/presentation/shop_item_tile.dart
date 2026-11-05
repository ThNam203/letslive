import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

/// Size of an item picture.
const double _pictureSize = 56;

/// One shop item as a row: picture, name, optional description, a [badge]
/// (price or quantity) and an optional [action] below. Shared by the shop,
/// the inventory, received gifts and the gift picker, like web's
/// `ShopItemCard`. A row (not web's grid card) because rows keep long names
/// and large text readable on a narrow phone.
///
/// With [onTap] the whole row is tappable (the gift picker).
class ShopItemTile extends StatelessWidget {
  const ShopItemTile({
    required this.name,
    required this.badge,
    this.imageUrl,
    this.description,
    this.action,
    this.onTap,
    super.key,
  });

  /// Item name.
  final String name;

  /// The picture; null when the item left the catalogue.
  final String? imageUrl;

  /// Short text under the name.
  final String? description;

  /// Price or quantity label.
  final String badge;

  /// Widget under the badge, e.g. the buy button.
  final Widget? action;

  /// Makes the row tappable; null leaves it inert.
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final text = description;
    final url = imageUrl;
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SizedBox.square(
                dimension: _pictureSize,
                child: url == null
                    ? const Icon(Icons.card_giftcard, size: _pictureSize / 2)
                    : CachedNetworkImage(
                        imageUrl: url,
                        fit: BoxFit.contain,
                        errorWidget: (_, _, _) => const Icon(
                          Icons.card_giftcard,
                          size: _pictureSize / 2,
                        ),
                      ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      name,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.titleSmall,
                    ),
                    if (text != null && text.isNotEmpty)
                      Text(
                        text,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    const SizedBox(height: 4),
                    Text(
                      badge,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.labelLarge,
                    ),
                    if (action != null) ...[const SizedBox(height: 8), action!],
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
