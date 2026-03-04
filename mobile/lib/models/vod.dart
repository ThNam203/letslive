class Vod {
  final String id;
  final String userId;
  final String? title;
  final String? description;
  final String? thumbnailUrl;
  final String? videoUrl;
  final int viewCount;
  final bool isPublic;
  final String createdAt;
  final String? updatedAt;

  // Joined user info
  final String? username;
  final String? displayName;
  final String? profilePicture;

  const Vod({
    required this.id,
    required this.userId,
    this.title,
    this.description,
    this.thumbnailUrl,
    this.videoUrl,
    this.viewCount = 0,
    this.isPublic = true,
    required this.createdAt,
    this.updatedAt,
    this.username,
    this.displayName,
    this.profilePicture,
  });

  factory Vod.fromJson(Map<String, dynamic> json) {
    return Vod(
      id: json['id'] as String,
      userId: json['userId'] as String,
      title: json['title'] as String?,
      description: json['description'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      videoUrl: json['videoUrl'] as String?,
      viewCount: json['viewCount'] as int? ?? 0,
      isPublic: json['isPublic'] as bool? ?? true,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String?,
      username: json['username'] as String?,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
    );
  }
}
