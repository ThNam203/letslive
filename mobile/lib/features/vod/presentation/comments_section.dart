import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';
import 'package:letslive_mobile/features/vod/comment_providers.dart';
import 'package:letslive_mobile/features/vod/presentation/action_feedback.dart';
import 'package:letslive_mobile/features/vod/presentation/comment_composer.dart';
import 'package:letslive_mobile/features/vod/presentation/comment_item.dart';

/// The comments under a video (web `CommentSection`): the title with the
/// total, a box to write a comment, the comments with their replies, and a
/// "Show more" button for the next page (web pages the same way).
class CommentsSection extends ConsumerWidget {
  const CommentsSection({
    required this.vodId,
    required this.ownerId,
    super.key,
  });

  final String vodId;

  /// The video's owner, for the "Owner" badge.
  final String ownerId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final comments = ref.watch(vodCommentsProvider(vodId));
    final notifier = ref.read(vodCommentsProvider(vodId).notifier);
    final feed = comments.value;
    // Whenever comments appear (first load, next page, a new comment), ask
    // which of them the viewer already liked. Comments already asked about
    // are skipped, so running this on every build costs nothing. Not during
    // build: it changes provider state.
    if (feed != null) {
      final likes = ref.read(commentLikesProvider(vodId).notifier);
      Future.microtask(() => likes.loadLiked(feed.items));
    }

    final total = feed?.total ?? 0;
    final status = feedStatusBody(
      value: comments.whenData((feed) => feed.items),
      onRetry: () => ref.invalidate(vodCommentsProvider(vodId)),
      empty: Padding(
        padding: const EdgeInsets.symmetric(vertical: 24),
        child: Text(
          l10n.comments_no_comments,
          textAlign: TextAlign.center,
          style: theme.textTheme.bodyMedium?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      ),
    );

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            total > 0 ? '${l10n.comments_title} ($total)' : l10n.comments_title,
            style: theme.textTheme.titleMedium,
          ),
          const SizedBox(height: 12),
          CommentComposer(
            hint: l10n.comments_write_comment,
            submitLabel: l10n.comments_post,
            onSubmit: (text) async {
              final failure = await notifier.post(text);
              if (context.mounted) showActionError(context, failure);
              return failure == null;
            },
          ),
          const SizedBox(height: 8),
          if (status != null)
            status
          else ...[
            for (final comment in feed!.items)
              Padding(
                padding: const EdgeInsets.only(top: 16),
                child: CommentItem(
                  key: ValueKey(comment.id),
                  comment: comment,
                  vodId: vodId,
                  ownerId: ownerId,
                  onEdit: (text) => notifier.edit(comment.id, text),
                  onDelete: () => notifier.delete(comment.id),
                ),
              ),
            if (feed.loadMoreError case final error?)
              FeedError(error: error, onRetry: notifier.loadMore)
            else if (feed.hasMore)
              Center(
                child: TextButton(
                  onPressed: feed.isLoadingMore ? null : notifier.loadMore,
                  child: Text(
                    feed.isLoadingMore
                        ? l10n.common_loading
                        : l10n.common_show_more,
                  ),
                ),
              ),
          ],
        ],
      ),
    );
  }
}
