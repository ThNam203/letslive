import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/home/domain/media_format.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';
import 'package:letslive_mobile/features/vod/comment_providers.dart';
import 'package:letslive_mobile/features/vod/data/vod_comment.dart';
import 'package:letslive_mobile/features/vod/presentation/comment_composer.dart';
import 'package:letslive_mobile/features/vod/presentation/timestamped_text.dart';

/// Avatar radius of a comment.
const double _avatarRadius = 16;

enum _Action { edit, delete }

/// One comment or reply and everything you can do with it (web
/// `CommentItem`): like, reply, view its replies, and, on your own comment
/// only, edit and delete.
///
/// [onEdit] and [onDelete] change the list that holds this comment (the
/// top-level list or its parent's replies); each returns the failed
/// response, or null.
class CommentItem extends ConsumerStatefulWidget {
  const CommentItem({
    required this.comment,
    required this.vodId,
    required this.ownerId,
    required this.onEdit,
    required this.onDelete,
    super.key,
  });

  /// The comment (or reply) to show.
  final VodComment comment;

  /// The video it belongs to.
  final String vodId;

  /// The video's owner, whose comments get an "Owner" badge.
  final String ownerId;

  /// Saves new [text] for this comment.
  final ActionResult Function(String text) onEdit;

  /// Deletes this comment.
  final ActionResult Function() onDelete;

  @override
  ConsumerState<CommentItem> createState() => _CommentItemState();
}

class _CommentItemState extends ConsumerState<CommentItem> {
  bool _isEditing = false;
  bool _showReplyForm = false;

  CommentKey get _key => (vodId: widget.vodId, commentId: widget.comment.id);

  Future<void> _confirmDelete() async {
    final l10n = AppLocalizations.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.comments_delete_confirm_title),
        content: Text(l10n.comments_delete_confirm_description),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.common_cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(l10n.comments_delete),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    final failure = await widget.onDelete();
    if (mounted) showActionError(context, failure);
  }

  Future<bool> _saveEdit(String text) async {
    final failure = await widget.onEdit(text);
    if (!mounted) return failure == null;
    showActionError(context, failure);
    if (failure == null) setState(() => _isEditing = false);
    return failure == null;
  }

  Future<bool> _postReply(String text) async {
    final failure = await ref
        .read(commentRepliesProvider(_key).notifier)
        .post(text);
    if (!mounted) return failure == null;
    showActionError(context, failure);
    if (failure == null) setState(() => _showReplyForm = false);
    return failure == null;
  }

  Future<void> _toggleLike() async {
    final failure = await ref
        .read(commentLikesProvider(widget.vodId).notifier)
        .toggle(widget.comment);
    if (mounted) showActionError(context, failure);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final comment = widget.comment;
    final author = comment.user;
    final me = ref.watch(sessionProvider).value?.user.id;
    final isMine = me == comment.userId;
    final likes = ref.watch(commentLikesProvider(widget.vodId));
    final replies = ref.watch(commentRepliesProvider(_key));
    final replyCount = comment.replyCount + replies.countDelta;
    final muted = theme.textTheme.bodySmall?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    final profile = author == null ? null : AppRoutes.channel(author.id);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Opacity(
          opacity: comment.isDeleted ? 0.6 : 1,
          child: InkWell(
            customBorder: const CircleBorder(),
            onTap: profile == null ? null : () => context.push(profile),
            child: UserAvatar(
              pictureUrl: author?.profilePicture,
              name: author?.username,
              radius: _avatarRadius,
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Opacity(
                opacity: comment.isDeleted ? 0.6 : 1,
                child: Wrap(
                  spacing: 8,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    InkWell(
                      onTap: profile == null
                          ? null
                          : () => context.push(profile),
                      child: Text(
                        author?.username ?? '...',
                        style: theme.textTheme.titleSmall,
                      ),
                    ),
                    if (comment.userId == widget.ownerId)
                      _Badge(l10n.comments_owner),
                    if (isMine) _Badge(l10n.comments_you),
                    Text(
                      timeAgo(l10n, comment.createdAt, DateTime.now()),
                      style: muted,
                    ),
                    if (comment.isEdited && !comment.isDeleted)
                      Text(
                        l10n.comments_edited,
                        style: muted?.copyWith(fontStyle: FontStyle.italic),
                      ),
                    if (isMine && !comment.isDeleted && !_isEditing)
                      PopupMenuButton<_Action>(
                        tooltip: l10n.comments_comment_actions,
                        padding: EdgeInsets.zero,
                        iconSize: 18,
                        onSelected: (action) => switch (action) {
                          _Action.edit => setState(() => _isEditing = true),
                          _Action.delete => _confirmDelete(),
                        },
                        itemBuilder: (context) => [
                          PopupMenuItem(
                            value: _Action.edit,
                            child: Text(l10n.comments_edit),
                          ),
                          PopupMenuItem(
                            value: _Action.delete,
                            child: Text(
                              l10n.comments_delete,
                              style: TextStyle(color: theme.colorScheme.error),
                            ),
                          ),
                        ],
                      ),
                  ],
                ),
              ),
              if (comment.isDeleted)
                Text(
                  l10n.comments_deleted_comment,
                  style: muted?.copyWith(fontStyle: FontStyle.italic),
                )
              else if (_isEditing)
                Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: CommentComposer(
                    hint: l10n.comments_edit_comment,
                    submitLabel: l10n.comments_save,
                    initialText: comment.content,
                    autofocus: true,
                    onSubmit: _saveEdit,
                    onCancel: () => setState(() => _isEditing = false),
                  ),
                )
              else
                TimestampedText(comment.content),
              Wrap(
                children: [
                  if (!comment.isDeleted && !_isEditing)
                    TextButton.icon(
                      key: ValueKey('like-${comment.id}'),
                      onPressed: _toggleLike,
                      icon: Icon(
                        likes.isLiked(comment.id)
                            ? Icons.favorite
                            : Icons.favorite_border,
                        size: 16,
                        color: likes.isLiked(comment.id) ? Colors.red : null,
                        semanticLabel: likes.isLiked(comment.id)
                            ? l10n.comments_unlike
                            : l10n.comments_like,
                      ),
                      label: Text(
                        likes.countOf(comment) > 0
                            ? '${likes.countOf(comment)}'
                            : '',
                      ),
                    ),
                  if (!_isEditing)
                    TextButton.icon(
                      onPressed: () =>
                          setState(() => _showReplyForm = !_showReplyForm),
                      icon: const Icon(Icons.reply, size: 16),
                      label: Text(l10n.comments_reply),
                    ),
                ],
              ),
              if (_showReplyForm)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: CommentComposer(
                    hint: l10n.comments_write_reply,
                    submitLabel: l10n.comments_post,
                    autofocus: true,
                    onSubmit: _postReply,
                    onCancel: () => setState(() => _showReplyForm = false),
                  ),
                ),
              if (replyCount > 0 && !replies.isShown)
                _loadRepliesButton(
                  replies,
                  l10n.comments_view_replies(replyCount),
                ),
              if (replies.isShown && replies.items.isNotEmpty)
                Container(
                  margin: const EdgeInsets.only(top: 8),
                  padding: const EdgeInsets.only(left: 12),
                  decoration: BoxDecoration(
                    border: Border(
                      left: BorderSide(color: theme.dividerColor, width: 2),
                    ),
                  ),
                  child: Column(
                    children: [
                      for (final reply in replies.items)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: CommentItem(
                            key: ValueKey(reply.id),
                            comment: reply,
                            vodId: widget.vodId,
                            ownerId: widget.ownerId,
                            onEdit: (text) => ref
                                .read(commentRepliesProvider(_key).notifier)
                                .edit(reply.id, text),
                            onDelete: () => ref
                                .read(commentRepliesProvider(_key).notifier)
                                .delete(reply.id),
                          ),
                        ),
                    ],
                  ),
                ),
              if (replies.isShown && replies.hasMore)
                _loadRepliesButton(replies, l10n.comments_load_more_replies),
              if (replies.error case final error?)
                Text(
                  apiErrorMessage(l10n, error.response),
                  style: muted?.copyWith(color: theme.colorScheme.error),
                ),
            ],
          ),
        ),
      ],
    );
  }

  /// "View n replies" / "Load more replies"; also the retry after an error.
  Widget _loadRepliesButton(RepliesState replies, String label) => Align(
    alignment: AlignmentDirectional.centerStart,
    child: TextButton(
      onPressed: replies.isLoading
          ? null
          : ref.read(commentRepliesProvider(_key).notifier).load,
      child: Text(
        replies.isLoading ? AppLocalizations.of(context).common_loading : label,
      ),
    ),
  );
}

/// "Owner" / "You" tag next to a name.
class _Badge extends StatelessWidget {
  const _Badge(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.primaryContainer,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
        child: Text(
          label,
          style: Theme.of(context).textTheme.labelSmall
              ?.copyWith(color: colors.onPrimaryContainer),
        ),
      ),
    );
  }
}
