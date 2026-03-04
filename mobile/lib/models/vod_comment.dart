class CommentUser {
  final String id;
  final String username;
  final String? displayName;
  final String? profilePicture;

  const CommentUser({
    required this.id,
    required this.username,
    this.displayName,
    this.profilePicture,
  });

  factory CommentUser.fromJson(Map<String, dynamic> json) {
    return CommentUser(
      id: json['id'] as String,
      username: json['username'] as String,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
    );
  }
}

class VodComment {
  final String id;
  final String vodId;
  final String userId;
  final String? parentId;
  final String content;
  final bool isDeleted;
  final int likeCount;
  final int replyCount;
  final String createdAt;
  final String updatedAt;
  final CommentUser? user;

  const VodComment({
    required this.id,
    required this.vodId,
    required this.userId,
    this.parentId,
    required this.content,
    this.isDeleted = false,
    this.likeCount = 0,
    this.replyCount = 0,
    required this.createdAt,
    required this.updatedAt,
    this.user,
  });

  factory VodComment.fromJson(Map<String, dynamic> json) {
    return VodComment(
      id: json['id'] as String,
      vodId: json['vodId'] as String,
      userId: json['userId'] as String,
      parentId: json['parentId'] as String?,
      content: json['content'] as String,
      isDeleted: json['isDeleted'] as bool? ?? false,
      likeCount: json['likeCount'] as int? ?? 0,
      replyCount: json['replyCount'] as int? ?? 0,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
      user: json['user'] != null
          ? CommentUser.fromJson(json['user'] as Map<String, dynamic>)
          : null,
    );
  }
}
