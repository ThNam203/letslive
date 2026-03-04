class Vod {
  final String id;
  final String livestreamId;
  final String userId;
  final String title;
  final String? description;
  final String? thumbnailUrl;
  final String visibility;
  final int viewCount;
  final int duration;
  final String playbackUrl;
  final String createdAt;
  final String updatedAt;

  // Joined user info (from backend joined queries)
  final String? username;
  final String? displayName;
  final String? profilePicture;

  const Vod({
    required this.id,
    required this.livestreamId,
    required this.userId,
    required this.title,
    this.description,
    this.thumbnailUrl,
    this.visibility = 'public',
    this.viewCount = 0,
    this.duration = 0,
    required this.playbackUrl,
    required this.createdAt,
    required this.updatedAt,
    this.username,
    this.displayName,
    this.profilePicture,
  });

  bool get isPublic => visibility == 'public';

  factory Vod.fromJson(Map<String, dynamic> json) {
    return Vod(
      id: json['id'] as String,
      livestreamId: json['livestreamId'] as String,
      userId: json['userId'] as String,
      title: json['title'] as String,
      description: json['description'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      visibility: json['visibility'] as String? ?? 'public',
      viewCount: json['viewCount'] as int? ?? 0,
      duration: json['duration'] as int? ?? 0,
      playbackUrl: json['playbackUrl'] as String,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
      username: json['username'] as String?,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
    );
  }
}
