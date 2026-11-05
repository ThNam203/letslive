/// The gift picker (web `GiftModal`) as a bottom sheet: pick a shop item and
/// it is bought and sent to the recipient in one request.
library;

import 'dart:async';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/shop/data/shop_models.dart';
import 'package:letslive_mobile/features/shop/presentation/shop_item_tile.dart';
import 'package:letslive_mobile/features/shop/shop_providers.dart';

/// How long the gift animation stays before the sheet closes (web: 3 s).
const Duration _animationDuration = Duration(seconds: 3);

/// Size of the gift animation (web: 256 px).
const double _animationSize = 256;

/// Opens the gift sheet for [recipientUserId], named [recipientName].
Future<void> showGiftSheet(
  BuildContext context, {
  required String recipientUserId,
  required String recipientName,
}) => showModalBottomSheet<void>(
  context: context,
  isScrollControlled: true,
  useSafeArea: true,
  builder: (_) =>
      GiftSheet(recipientUserId: recipientUserId, recipientName: recipientName),
);

/// Content of the gift sheet. While a send is in flight every item is
/// disabled (`purchaseProvider`); a failure is shown in the sheet and the
/// balance reloads, nothing is retried by itself.
class GiftSheet extends ConsumerStatefulWidget {
  const GiftSheet({
    required this.recipientUserId,
    required this.recipientName,
    super.key,
  });

  /// Who gets the gift.
  final String recipientUserId;

  /// Shown in the title.
  final String recipientName;

  @override
  ConsumerState<GiftSheet> createState() => _GiftSheetState();
}

class _GiftSheetState extends ConsumerState<GiftSheet> {
  /// Result line under the title: "Gift sent!" or the error.
  String? _status;
  bool _statusIsError = false;

  /// Animation of the gift just sent, shown over the sheet.
  String? _animationUrl;
  Timer? _closeTimer;

  @override
  void dispose() {
    _closeTimer?.cancel();
    super.dispose();
  }

  Future<void> _send(ShopItem item) async {
    final l10n = AppLocalizations.of(context);
    final response = await ref
        .read(purchaseProvider.notifier)
        .sendGift(item.id, widget.recipientUserId);
    if (response == null || !mounted) return;
    setState(() {
      _statusIsError = !response.success;
      _status = response.success
          ? l10n.shop_shop_gift_sent
          : apiErrorMessage(l10n, response);
      final url = response.data?.animationUrl ?? '';
      if (response.success && url.isNotEmpty) {
        _animationUrl = url;
        _closeTimer = Timer(_animationDuration, _close);
      }
    });
  }

  void _close() {
    if (mounted) Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final items = ref.watch(shopItemsProvider);
    final pendingId = ref.watch(purchaseProvider);
    final status = feedStatusBody(
      value: items,
      onRetry: () => ref.invalidate(shopItemsProvider),
      empty: EmptyNotice(
        icon: Icons.card_giftcard,
        title: l10n.shop_shop_gift_no_items,
      ),
    );
    final statusText = _status;
    final animation = _animationUrl;
    return Stack(
      children: [
        Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text(
                '${l10n.shop_shop_gift_pick_item} — ${widget.recipientName}',
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.titleMedium,
              ),
            ),
            if (statusText != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Text(
                  statusText,
                  key: const ValueKey('gift-status'),
                  style: TextStyle(
                    color: _statusIsError
                        ? theme.colorScheme.error
                        : theme.colorScheme.primary,
                  ),
                ),
              ),
            Flexible(
              child: status != null
                  ? ListView(shrinkWrap: true, children: [status])
                  : ListView(
                      shrinkWrap: true,
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      children: [
                        for (final item in items.requireValue)
                          ShopItemTile(
                            key: ValueKey('gift-item-${item.id}'),
                            name: item.name,
                            imageUrl: item.imageUrl,
                            badge: pendingId == item.id
                                ? l10n.shop_shop_gift_sending
                                : l10n.shop_shop_price_label(item.price),
                            onTap: pendingId == null ? () => _send(item) : null,
                          ),
                      ],
                    ),
            ),
          ],
        ),
        if (animation != null)
          Positioned.fill(
            child: GestureDetector(
              key: const ValueKey('gift-animation'),
              onTap: _close,
              child: ColoredBox(
                color: Colors.black87,
                child: Center(
                  child: CachedNetworkImage(
                    imageUrl: animation,
                    width: _animationSize,
                    height: _animationSize,
                    fit: BoxFit.contain,
                    errorWidget: (_, _, _) => const Icon(
                      Icons.card_giftcard,
                      size: _animationSize / 2,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}
