import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/shop/presentation/gift_sheet.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';

/// The "send a gift" icon next to the follow button (web `ProfileActions`).
/// Hidden until the user has loaded and on the signed-in user's own profile.
/// Opens the gift sheet.
class GiftButton extends ConsumerWidget {
  const GiftButton({required this.userId, super.key});

  /// The user who would get the gift.
  final String userId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final myId = ref.watch(sessionProvider).value?.user.id;
    final user = ref.watch(publicUserProvider(userId)).value;
    if (user == null || myId == null || myId == userId) {
      return const SizedBox.shrink();
    }
    return IconButton.outlined(
      key: const ValueKey('gift-button'),
      tooltip: l10n.shop_shop_gift_button,
      icon: const Icon(Icons.card_giftcard),
      // Same height as the follow button next to it (keeps the header's
      // height budget, see `profileHeaderHeight`).
      style: IconButton.styleFrom(
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      ),
      onPressed: () => showGiftSheet(
        context,
        recipientUserId: userId,
        recipientName: user.username,
      ),
    );
  }
}
