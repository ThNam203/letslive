import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/vod/data/vod_reaction.dart';
import 'package:letslive_mobile/features/vod/presentation/action_feedback.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';

/// Like and dislike buttons with the like count (web
/// `VODReactionButtons`). Until the viewer's own reaction has loaded (or if
/// it failed to) the buttons show the video's like count and no reaction,
/// like web. A tap shows at once and is undone with a message if refused.
class VodReactionButtons extends ConsumerWidget {
  const VodReactionButtons({required this.vod, super.key});

  final Vod vod;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final current =
        ref.watch(vodReactionProvider(vod.id)).value ??
        VodReactionState(likeCount: vod.likeCount);
    final reaction = current.reaction;

    Future<void> react(VodReaction clicked) async {
      final failure = await ref
          .read(vodReactionProvider(vod.id).notifier)
          .react(clicked, current);
      if (context.mounted) showActionError(context, failure);
    }

    return DecoratedBox(
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(24),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextButton.icon(
            key: const ValueKey('vod-like'),
            onPressed: () => react(VodReaction.like),
            icon: Icon(
              reaction == VodReaction.like
                  ? Icons.thumb_up
                  : Icons.thumb_up_outlined,
              semanticLabel: reaction == VodReaction.like
                  ? l10n.common_vod_remove_like
                  : l10n.common_vod_like,
            ),
            label: Text(compactCount(l10n, current.likeCount)),
          ),
          IconButton(
            key: const ValueKey('vod-dislike'),
            tooltip: reaction == VodReaction.dislike
                ? l10n.common_vod_remove_dislike
                : l10n.common_vod_dislike,
            onPressed: () => react(VodReaction.dislike),
            icon: Icon(
              reaction == VodReaction.dislike
                  ? Icons.thumb_down
                  : Icons.thumb_down_outlined,
            ),
          ),
        ],
      ),
    );
  }
}
