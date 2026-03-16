import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../core/config/app_config.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/vod_comment.dart';
import '../../../providers.dart';

class VodCommentSection extends ConsumerStatefulWidget {
  final String vodId;
  final String? vodOwnerId;

  const VodCommentSection({
    super.key,
    required this.vodId,
    this.vodOwnerId,
  });

  @override
  ConsumerState<VodCommentSection> createState() => _VodCommentSectionState();
}

class _VodCommentSectionState extends ConsumerState<VodCommentSection> {
  final _commentController = TextEditingController();
  List<VodComment> _comments = [];
  Set<String> _likedIds = {};
  bool _isLoading = true;
  bool _isLoadingMore = false;
  bool _isPosting = false;
  String? _error;
  int _currentPage = 0;
  bool _hasMore = true;

  @override
  void initState() {
    super.initState();
    _fetchComments();
  }

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  Future<void> _fetchComments() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _currentPage = 0;
    });

    try {
      final repo = ref.read(vodCommentRepositoryProvider);
      final response = await repo.getComments(widget.vodId, page: 0);
      if (!mounted) return;

      if (response.success) {
        final comments = response.data ?? [];
        final pageSize = response.meta?.pageSize ?? 10;
        final total = response.meta?.total ?? 0;
        setState(() {
          _comments = comments;
          _isLoading = false;
          _hasMore = (comments.length >= pageSize) && (comments.length < total);
        });
        _fetchLikedIds(comments);
      } else {
        setState(() {
          _error = response.message;
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = AppLocalizations.of(context).fetchError;
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMore() async {
    if (_isLoadingMore || !_hasMore) return;
    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(vodCommentRepositoryProvider);
      final nextPage = _currentPage + 1;
      final response =
          await repo.getComments(widget.vodId, page: nextPage);
      if (!mounted) return;

      if (response.success) {
        final newComments = response.data ?? [];
        final pageSize = response.meta?.pageSize ?? 10;
        final total = response.meta?.total ?? 0;
        setState(() {
          _currentPage = nextPage;
          _comments.addAll(newComments);
          _isLoadingMore = false;
          _hasMore = (_comments.length < total) && (newComments.length >= pageSize);
        });
        _fetchLikedIds(newComments);
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  Future<void> _fetchLikedIds(List<VodComment> comments) async {
    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null || comments.isEmpty) return;

    try {
      final repo = ref.read(vodCommentRepositoryProvider);
      final ids = comments.map((c) => c.id).toList();
      final response = await repo.getLikedCommentIds(ids);
      if (mounted && response.success && response.data != null) {
        setState(() => _likedIds.addAll(response.data!));
      }
    } catch (_) {}
  }

  Future<void> _postComment() async {
    final text = _commentController.text.trim();
    if (text.isEmpty) return;

    setState(() => _isPosting = true);

    try {
      final repo = ref.read(vodCommentRepositoryProvider);
      final response =
          await repo.createComment(widget.vodId, content: text);
      if (!mounted) return;

      if (response.success && response.data != null) {
        setState(() {
          _comments.insert(0, response.data!);
          _isPosting = false;
        });
        _commentController.clear();
      } else {
        setState(() => _isPosting = false);
        _showError(response.message);
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() => _isPosting = false);
        _showError(AppLocalizations.of(context).fetchError);
      }
    }
  }

  Future<void> _deleteComment(VodComment comment) async {
    final l10n = AppLocalizations.of(context);
    final confirmed = await showAdaptiveDialog<bool>(
      context: context,
      builder: (context) => AlertDialog.adaptive(
        title: Text(l10n.commentsDeleteConfirmTitle),
        content: Text(l10n.commentsDeleteConfirmDescription),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(l10n.delete,
                style: TextStyle(color: context.theme.colors.destructive)),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    try {
      final repo = ref.read(vodCommentRepositoryProvider);
      final response = await repo.deleteComment(comment.id);
      if (!mounted) return;

      if (response.success) {
        setState(() => _comments.removeWhere((c) => c.id == comment.id));
      } else {
        _showError(response.message);
      }
    } on DioException catch (_) {
      if (mounted) _showError(AppLocalizations.of(context).fetchError);
    }
  }

  Future<void> _toggleLike(VodComment comment) async {
    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) return;

    final isLiked = _likedIds.contains(comment.id);
    final repo = ref.read(vodCommentRepositoryProvider);

    // Optimistic update
    setState(() {
      if (isLiked) {
        _likedIds.remove(comment.id);
      } else {
        _likedIds.add(comment.id);
      }
      final idx = _comments.indexWhere((c) => c.id == comment.id);
      if (idx != -1) {
        final old = _comments[idx];
        _comments[idx] = VodComment(
          id: old.id,
          vodId: old.vodId,
          userId: old.userId,
          parentId: old.parentId,
          content: old.content,
          isDeleted: old.isDeleted,
          likeCount: old.likeCount + (isLiked ? -1 : 1),
          replyCount: old.replyCount,
          createdAt: old.createdAt,
          updatedAt: old.updatedAt,
          user: old.user,
        );
      }
    });

    try {
      final response = isLiked
          ? await repo.unlikeComment(comment.id)
          : await repo.likeComment(comment.id);

      if (!response.success) {
        // Revert on failure
        if (mounted) {
          setState(() {
            if (isLiked) {
              _likedIds.add(comment.id);
            } else {
              _likedIds.remove(comment.id);
            }
          });
        }
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          if (isLiked) {
            _likedIds.add(comment.id);
          } else {
            _likedIds.remove(comment.id);
          }
        });
      }
    }
  }

  void _showError(String message) {
    showFToast(
      context: context,
      title: Text(message),
      icon: const Icon(FIcons.circleAlert),
    );
  }

  String _formatTimeAgo(String createdAt, AppLocalizations l10n) {
    final date = DateTime.tryParse(createdAt);
    if (date == null) return '';

    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inSeconds < 60) return l10n.timeSecondsAgo(diff.inSeconds);
    if (diff.inMinutes < 60) return l10n.timeMinutesAgo(diff.inMinutes);
    if (diff.inHours < 24) return l10n.timeHoursAgo(diff.inHours);
    if (diff.inDays < 7) return l10n.timeDaysAgo(diff.inDays);
    if (diff.inDays < 30) return l10n.timeWeeksAgo(diff.inDays ~/ 7);
    if (diff.inDays < 365) return l10n.timeMonthsAgo(diff.inDays ~/ 30);
    return l10n.timeYearsAgo(diff.inDays ~/ 365);
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);
    final currentUser = ref.watch(currentUserProvider);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Divider(height: 1),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
          child: Text(
            l10n.commentsTitle,
            style: typography.base.copyWith(fontWeight: FontWeight.bold),
          ),
        ),

        // Comment input
        if (currentUser != null)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Expanded(
                  child: TextField(
                    controller: _commentController,
                    maxLines: 3,
                    minLines: 1,
                    decoration: InputDecoration(
                      hintText: l10n.commentsWriteComment,
                      hintStyle: typography.sm.copyWith(
                        color: colors.mutedForeground,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: colors.border),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: colors.border),
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 10,
                      ),
                      isDense: true,
                    ),
                    style: typography.sm,
                  ),
                ),
                const SizedBox(width: 8),
                FButton(
                  onPress: _isPosting ? null : _postComment,
                  child: _isPosting
                      ? const SizedBox(
                          height: 16,
                          width: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.commentsPost),
                ),
              ],
            ),
          )
        else
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Text(
              l10n.commentsLoginToComment,
              style: typography.sm.copyWith(color: colors.mutedForeground),
            ),
          ),

        // Comment list
        if (_isLoading)
          const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: FCircularProgress()),
          )
        else if (_error != null)
          Padding(
            padding: const EdgeInsets.all(16),
            child: Center(
              child: Column(
                children: [
                  Text(_error!,
                      style:
                          typography.sm.copyWith(color: colors.mutedForeground)),
                  const SizedBox(height: 8),
                  FButton(
                    variant: FButtonVariant.outline,
                    onPress: _fetchComments,
                    child: Text(l10n.retry),
                  ),
                ],
              ),
            ),
          )
        else if (_comments.isEmpty)
          Padding(
            padding: const EdgeInsets.all(24),
            child: Center(
              child: Text(
                l10n.commentsNoComments,
                style: typography.sm.copyWith(color: colors.mutedForeground),
              ),
            ),
          )
        else ...[
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _comments.length,
            separatorBuilder: (_, _) =>
                Divider(height: 1, color: colors.border),
            itemBuilder: (context, index) {
              final comment = _comments[index];
              return _CommentTile(
                comment: comment,
                isLiked: _likedIds.contains(comment.id),
                isOwner: currentUser?.id == comment.userId,
                isVodOwner: widget.vodOwnerId == currentUser?.id,
                timeAgo: _formatTimeAgo(comment.createdAt, l10n),
                onLike: () => _toggleLike(comment),
                onDelete: () => _deleteComment(comment),
              );
            },
          ),
          if (_hasMore)
            Padding(
              padding: const EdgeInsets.all(16),
              child: Center(
                child: FButton(
                  variant: FButtonVariant.outline,
                  onPress: _isLoadingMore ? null : _loadMore,
                  child: _isLoadingMore
                      ? const SizedBox(
                          height: 16,
                          width: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.messagesLoadMore),
                ),
              ),
            ),
        ],
      ],
    );
  }
}

class _CommentTile extends StatelessWidget {
  final VodComment comment;
  final bool isLiked;
  final bool isOwner;
  final bool isVodOwner;
  final String timeAgo;
  final VoidCallback onLike;
  final VoidCallback onDelete;

  const _CommentTile({
    required this.comment,
    required this.isLiked,
    required this.isOwner,
    required this.isVodOwner,
    required this.timeAgo,
    required this.onLike,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    if (comment.isDeleted) {
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Text(
          l10n.commentsDeletedComment,
          style: typography.sm
              .copyWith(color: colors.mutedForeground, fontStyle: FontStyle.italic),
        ),
      );
    }

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CircleAvatar(
            radius: 16,
            backgroundImage: comment.user?.profilePicture != null
                ? CachedNetworkImageProvider(
                    '${AppConfig.apiUrl}/${comment.user!.profilePicture}')
                : null,
            child: comment.user?.profilePicture == null
                ? const Icon(FIcons.user, size: 16)
                : null,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      comment.user?.displayName ??
                          comment.user?.username ??
                          '',
                      style: typography.xs
                          .copyWith(fontWeight: FontWeight.w600),
                    ),
                    if (comment.userId == comment.vodId) ...[
                      const SizedBox(width: 4),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 4, vertical: 1),
                        decoration: BoxDecoration(
                          color: colors.primary.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          l10n.commentsOwner,
                          style: typography.xs.copyWith(
                            color: colors.primary,
                            fontSize: 10,
                          ),
                        ),
                      ),
                    ],
                    const Spacer(),
                    Text(
                      timeAgo,
                      style:
                          typography.xs.copyWith(color: colors.mutedForeground),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Text(comment.content, style: typography.sm),
                const SizedBox(height: 8),
                Row(
                  children: [
                    GestureDetector(
                      onTap: onLike,
                      child: Row(
                        children: [
                          Icon(
                            isLiked ? FIcons.heartCrack : FIcons.heart,
                            size: 14,
                            color: isLiked
                                ? colors.destructive
                                : colors.mutedForeground,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            '${comment.likeCount}',
                            style: typography.xs.copyWith(
                              color: isLiked
                                  ? colors.destructive
                                  : colors.mutedForeground,
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (isOwner || isVodOwner) ...[
                      const SizedBox(width: 16),
                      GestureDetector(
                        onTap: onDelete,
                        child: Row(
                          children: [
                            Icon(FIcons.trash,
                                size: 14, color: colors.mutedForeground),
                            const SizedBox(width: 4),
                            Text(
                              l10n.commentsDelete,
                              style: typography.xs
                                  .copyWith(color: colors.mutedForeground),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
