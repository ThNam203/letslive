/// Riverpod providers for the comments under a video, following web's
/// `use-vod-comments.ts` and `vod-comments-cache.ts`:
///
/// - [vodCommentsProvider]: top-level comments, paged, with the running
///   total; posting, editing and deleting update the loaded list.
/// - [commentRepliesProvider]: the replies of one comment, loaded on demand.
/// - [commentLikesProvider]: which comments the viewer liked and the like
///   counts they changed; likes are optimistic and undone on failure.
///
/// Every action returns the failed response (null on success) so the screen
/// can show the localized reason. All providers are `autoDispose`.
library;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/vod/data/vod_comment.dart';
import 'package:letslive_mobile/features/vod/data/vod_comment_repository.dart';

/// Comment endpoints.
final vodCommentRepositoryProvider = Provider<VodCommentRepository>(
  (ref) => VodCommentRepository(ref.watch(apiClientProvider)),
);

/// A failed action's response, or null when it worked.
typedef ActionResult = Future<ApiResponse<Object?>?>;

/// The signed-in user as a comment author (the server's answer to a create
/// has no author).
CommentUser? _currentAuthor(Ref ref) {
  final me = ref.read(sessionProvider).value?.user;
  if (me == null) return null;
  return CommentUser(
    id: me.id,
    username: me.username,
    profilePicture: me.profilePicture,
  );
}

VodComment _withAuthor(Ref ref, VodComment created) {
  final author = _currentAuthor(ref);
  return created.user == null && author != null && author.id == created.userId
      ? created.withUser(author)
      : created;
}

/// The comments of a video.
final vodCommentsProvider = AsyncNotifierProvider.autoDispose
    .family<VodCommentsNotifier, PagedFeed<VodComment>, String>(
      VodCommentsNotifier.new,
      retry: noRetry,
    );

/// Top-level comments of [vodId], newest first, [commentsPageSize] per page.
class VodCommentsNotifier extends PagedFeedNotifier<VodComment> {
  VodCommentsNotifier(this.vodId);

  /// The commented video.
  final String vodId;

  @override
  int get pageSize => commentsPageSize;

  @override
  String idOf(VodComment item) => item.id;

  @override
  Future<ApiResponse<List<VodComment>>> fetchPage(int page) =>
      ref.read(vodCommentRepositoryProvider).list(vodId, page: page);

  /// Posts a comment; on success it goes to the top and the total grows.
  ActionResult post(String content) async {
    final response = await ref
        .read(vodCommentRepositoryProvider)
        .create(vodId, content);
    final created = response.data;
    final feed = state.value;
    if (!ref.mounted || feed == null) return null;
    if (!response.success || created == null) return response;
    state = AsyncData(
      feed.copyWith(
        items: [_withAuthor(ref, created), ...feed.items],
        total: (feed.total ?? feed.items.length) + 1,
      ),
    );
    return null;
  }

  /// Replaces the text of comment [id].
  ActionResult edit(String id, String content) async {
    final response = await ref
        .read(vodCommentRepositoryProvider)
        .edit(id, content);
    if (!ref.mounted) return null;
    if (!response.success) return response;
    _replace(id, (comment) => comment.edited(content));
    return null;
  }

  /// Deletes comment [id]; it stays in the list as "deleted" (its replies
  /// remain) and the total shrinks.
  ActionResult delete(String id) async {
    final response = await ref.read(vodCommentRepositoryProvider).delete(id);
    final feed = state.value;
    if (!ref.mounted || feed == null) return null;
    if (!response.success) return response;
    state = AsyncData(
      feed.copyWith(
        items: [for (final c in feed.items) c.id == id ? c.deleted() : c],
        total: ((feed.total ?? feed.items.length) - 1).clamp(0, 1 << 30),
      ),
    );
    return null;
  }

  void _replace(String id, VodComment Function(VodComment) change) {
    final feed = state.value;
    if (feed == null) return;
    state = AsyncData(
      feed.copyWith(
        items: [for (final c in feed.items) c.id == id ? change(c) : c],
      ),
    );
  }
}

/// Identifies the replies of one comment.
typedef CommentKey = ({String vodId, String commentId});

/// What is loaded of one comment's replies.
class RepliesState {
  const RepliesState({
    this.items = const [],
    this.isShown = false,
    this.hasMore = true,
    this.isLoading = false,
    this.nextPage = 0,
    this.countDelta = 0,
    this.error,
  });

  /// Replies loaded so far, oldest first.
  final List<VodComment> items;

  /// Whether the replies were requested (the "view replies" button is gone).
  final bool isShown;

  /// Whether another page may exist.
  final bool hasMore;
  final bool isLoading;
  final int nextPage;

  /// Replies added minus deleted here since the comment was loaded; web
  /// keeps the same running `replyCount`.
  final int countDelta;

  /// The last page request failed.
  final ApiLoadException? error;

  RepliesState copyWith({
    List<VodComment>? items,
    bool? isShown,
    bool? hasMore,
    bool? isLoading,
    int? nextPage,
    int? countDelta,
    ApiLoadException? error,
    bool clearError = false,
  }) => RepliesState(
    items: items ?? this.items,
    isShown: isShown ?? this.isShown,
    hasMore: hasMore ?? this.hasMore,
    isLoading: isLoading ?? this.isLoading,
    nextPage: nextPage ?? this.nextPage,
    countDelta: countDelta ?? this.countDelta,
    error: clearError ? null : error ?? this.error,
  );
}

/// The replies of a comment.
final commentRepliesProvider = NotifierProvider.autoDispose
    .family<CommentRepliesNotifier, RepliesState, CommentKey>(
      CommentRepliesNotifier.new,
    );

/// Replies of one comment: loaded page by page on demand, then changed
/// locally by posting, editing and deleting (web keeps them in the comment
/// item's own state the same way).
class CommentRepliesNotifier extends Notifier<RepliesState> {
  CommentRepliesNotifier(this.key);

  /// The comment whose replies these are.
  final CommentKey key;

  @override
  RepliesState build() => const RepliesState();

  /// Loads the first page, or the next one (also the retry after an error).
  Future<void> load() async {
    if (state.isLoading || !state.hasMore) return;
    state = state.copyWith(isLoading: true, clearError: true);
    final response = await ref
        .read(vodCommentRepositoryProvider)
        .replies(key.commentId, page: state.nextPage);
    if (!ref.mounted) return;
    if (!response.success) {
      state = state.copyWith(
        isLoading: false,
        error: ApiLoadException(response),
      );
      return;
    }
    final page = response.data ?? const <VodComment>[];
    final known = {for (final r in state.items) r.id};
    state = state.copyWith(
      items: [
        ...state.items,
        for (final r in page)
          if (!known.contains(r.id)) r,
      ],
      isShown: true,
      isLoading: false,
      hasMore: page.length >= repliesPageSize,
      nextPage: state.nextPage + 1,
    );
    await ref.read(commentLikesProvider(key.vodId).notifier).loadLiked(page);
  }

  /// Posts a reply; it goes to the end and the reply count grows.
  ActionResult post(String content) async {
    final response = await ref
        .read(vodCommentRepositoryProvider)
        .create(key.vodId, content, parentId: key.commentId);
    final created = response.data;
    if (!ref.mounted) return null;
    if (!response.success || created == null) return response;
    state = state.copyWith(
      items: [...state.items, _withAuthor(ref, created)],
      isShown: true,
      countDelta: state.countDelta + 1,
    );
    return null;
  }

  /// Replaces the text of reply [id].
  ActionResult edit(String id, String content) async {
    final response = await ref
        .read(vodCommentRepositoryProvider)
        .edit(id, content);
    if (!ref.mounted) return null;
    if (!response.success) return response;
    state = state.copyWith(
      items: [for (final r in state.items) r.id == id ? r.edited(content) : r],
    );
    return null;
  }

  /// Deletes reply [id]; it stays as "deleted" and the reply count shrinks.
  ActionResult delete(String id) async {
    final response = await ref.read(vodCommentRepositoryProvider).delete(id);
    if (!ref.mounted) return null;
    if (!response.success) return response;
    state = state.copyWith(
      items: [for (final r in state.items) r.id == id ? r.deleted() : r],
      countDelta: state.countDelta - 1,
    );
    return null;
  }
}

/// The viewer's likes among the comments of one video.
class CommentLikes {
  const CommentLikes({this.liked = const {}, this.counts = const {}});

  /// Ids of the comments the viewer liked.
  final Set<String> liked;

  /// Like counts changed by the viewer, by comment id; the comment's own
  /// count is used for the rest.
  final Map<String, int> counts;

  /// Whether the viewer liked [commentId].
  bool isLiked(String commentId) => liked.contains(commentId);

  /// The like count to show for [comment].
  int countOf(VodComment comment) => counts[comment.id] ?? comment.likeCount;
}

/// Likes of the comments of a video.
final commentLikesProvider = NotifierProvider.autoDispose
    .family<CommentLikesNotifier, CommentLikes, String>(
      CommentLikesNotifier.new,
    );

/// Loads which comments the viewer liked and toggles likes optimistically.
class CommentLikesNotifier extends Notifier<CommentLikes> {
  CommentLikesNotifier(this.vodId);

  /// The commented video.
  final String vodId;

  final Set<String> _requested = {};
  final Set<String> _pending = {};

  @override
  CommentLikes build() => const CommentLikes();

  /// Asks which of [comments] the viewer liked, once per comment. A failed
  /// request leaves them unliked and is retried the next time they are
  /// passed in (when the list changes).
  Future<void> loadLiked(List<VodComment> comments) async {
    final ids = [
      for (final c in comments)
        if (!c.isDeleted && _requested.add(c.id)) c.id,
    ];
    if (ids.isEmpty) return;
    final response = await ref.read(vodCommentRepositoryProvider).likedIds(ids);
    if (!ref.mounted) return;
    final liked = response.data;
    if (!response.success || liked == null) {
      _requested.removeAll(ids);
      return;
    }
    state = CommentLikes(
      liked: {...state.liked, ...liked},
      counts: state.counts,
    );
  }

  /// Likes or unlikes [comment] at once and undoes it if the server refuses.
  /// Ignored while the same comment's request is in flight (like web).
  ActionResult toggle(VodComment comment) async {
    final id = comment.id;
    if (!_pending.add(id)) return null;
    final before = state;
    final wasLiked = before.isLiked(id);
    final count = before.countOf(comment);
    state = CommentLikes(
      liked: wasLiked ? ({...before.liked}..remove(id)) : {...before.liked, id},
      counts: {
        ...before.counts,
        id: (count + (wasLiked ? -1 : 1)).clamp(0, 1 << 30),
      },
    );
    final repo = ref.read(vodCommentRepositoryProvider);
    final response = await (wasLiked ? repo.unlike(id) : repo.like(id));
    _pending.remove(id);
    if (!ref.mounted) return null;
    if (response.success) return null;
    // Undo only this comment; other likes may have changed meanwhile.
    state = CommentLikes(
      liked: wasLiked ? {...state.liked, id} : ({...state.liked}..remove(id)),
      counts: {...state.counts, id: count},
    );
    return response;
  }
}
