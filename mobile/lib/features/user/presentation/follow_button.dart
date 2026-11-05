import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/user/user_providers.dart';
import 'package:letslive_mobile/features/vod/presentation/action_feedback.dart';

/// Follow / Unfollow button for [userId] (web `ProfileActions`). Hidden until
/// the user has loaded and for the signed-in user's own profile. The change
/// is optimistic with a rollback and a message on failure; see
/// `PublicUserNotifier.toggleFollow`.
class FollowButton extends ConsumerWidget {
  const FollowButton({required this.userId, super.key});

  /// The user to follow.
  final String userId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final myId = ref.watch(sessionProvider).value?.user.id;
    final user = ref.watch(publicUserProvider(userId)).value;
    if (user == null || myId == null || myId == userId) {
      return const SizedBox.shrink();
    }
    final following = user.isFollowing ?? false;
    Future<void> toggle() async {
      final failure = await ref
          .read(publicUserProvider(userId).notifier)
          .toggleFollow();
      if (context.mounted) showActionError(context, failure);
    }

    return following
        ? OutlinedButton(
            key: const ValueKey('follow-button'),
            onPressed: toggle,
            child: Text(l10n.common_unfollow),
          )
        : FilledButton(
            key: const ValueKey('follow-button'),
            onPressed: toggle,
            child: Text(l10n.common_follow),
          );
  }
}
