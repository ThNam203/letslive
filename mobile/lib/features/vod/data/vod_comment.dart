import 'package:letslive_mobile/core/network/json_fields.dart';

/// Who wrote a comment (`CommentUser` in `web/types/vod-comment.ts`).
class CommentUser {
  const CommentUser({
    required this.id,
    required this.username,
    this.profilePicture,
  });

  factory CommentUser.fromJson(Map<String, Object?> json) {
    final f = JsonFields('CommentUser', json);
    return CommentUser(
      id: f.string('id'),
      username: f.string('username'),
      profilePicture: f.optString('profilePicture'),
    );
  }

  /// The author's user id.
  final String id;
  final String username;

  /// Absolute URL of the avatar; null when none was uploaded.
  final String? profilePicture;
}

/// A comment on a video, or a reply to one, typed after `VODComment` in
/// `web/types/vod-comment.ts` and the Go `dto.VODCommentWithUser`.
///
/// Immutable: edits and deletes make a changed copy.
class VodComment {
  const VodComment({
    required this.id,
    required this.vodId,
    required this.userId,
    required this.content,
    required this.isDeleted,
    required this.isEdited,
    required this.likeCount,
    required this.replyCount,
    required this.createdAt,
    this.parentId,
    this.user,
  });

  factory VodComment.fromJson(Map<String, Object?> json) {
    final f = JsonFields('VodComment', json);
    final user = json['user'];
    return VodComment(
      id: f.string('id'),
      vodId: f.string('vodId'),
      userId: f.string('userId'),
      parentId: f.optString('parentId'),
      content: f.string('content'),
      isDeleted: f.boolean('isDeleted'),
      isEdited: f.boolean('isEdited'),
      likeCount: f.integer('likeCount'),
      replyCount: f.integer('replyCount'),
      createdAt: f.date('createdAt'),
      // Best effort on the server: a comment may come without its author.
      user: user == null
          ? null
          : parseJsonObject(user, 'CommentUser', CommentUser.fromJson),
    );
  }

  /// `ParseData<VodComment>` for `ApiClient`.
  static VodComment parse(Object? data) =>
      parseJsonObject(data, 'VodComment', VodComment.fromJson);

  /// `ParseData<List<VodComment>>` for `ApiClient`.
  static List<VodComment> parseList(Object? data) =>
      parseJsonList(data, 'VodComment', VodComment.fromJson);

  /// The comment's own id.
  final String id;

  /// The video the comment is on.
  final String vodId;

  /// The author's user id.
  final String userId;

  /// The comment this one replies to; null for a top-level comment.
  final String? parentId;

  /// Empty once [isDeleted] (the server strips it).
  final String content;
  final bool isDeleted;
  final bool isEdited;
  final int likeCount;
  final int replyCount;
  final DateTime createdAt;

  /// The author; null when the server could not look the user up.
  final CommentUser? user;

  /// This comment with [content] replaced and marked edited.
  VodComment edited(String content) => _copy(content: content, isEdited: true);

  /// This comment as the server shows a deleted one: no text.
  VodComment deleted() => _copy(content: '', isDeleted: true);

  /// This comment with [user] set (a created comment comes without one).
  VodComment withUser(CommentUser user) => _copy(user: user);

  VodComment _copy({
    String? content,
    bool? isDeleted,
    bool? isEdited,
    CommentUser? user,
  }) => VodComment(
    id: id,
    vodId: vodId,
    userId: userId,
    parentId: parentId,
    content: content ?? this.content,
    isDeleted: isDeleted ?? this.isDeleted,
    isEdited: isEdited ?? this.isEdited,
    likeCount: likeCount,
    replyCount: replyCount,
    createdAt: createdAt,
    user: user ?? this.user,
  );
}
