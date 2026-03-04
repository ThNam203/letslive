class Livestream {
  final String id;
  final String userId;
  final String title;
  final String? description;
  final String? thumbnailUrl;
  final int viewCount;
  final String visibility;
  final String startedAt;
  final String? endedAt;
  final String createdAt;
  final String updatedAt;
  final String? vodId;

  // Joined user info (from backend joined queries)
  final String? username;
  final String? displayName;
  final String? profilePicture;

  const Livestream({
    required this.id,
    required this.userId,
    required this.title,
    this.description,
    this.thumbnailUrl,
    this.viewCount = 0,
    this.visibility = 'public',
    required this.startedAt,
    this.endedAt,
    required this.createdAt,
    required this.updatedAt,
    this.vodId,
    this.username,
    this.displayName,
    this.profilePicture,
  });

  bool get isLive => endedAt == null;

  factory Livestream.fromJson(Map<String, dynamic> json) {
    return Livestream(
      id: json['id'] as String,
      userId: json['userId'] as String,
      title: json['title'] as String,
      description: json['description'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      viewCount: json['viewCount'] as int? ?? 0,
      visibility: json['visibility'] as String? ?? 'public',
      startedAt: json['startedAt'] as String,
      endedAt: json['endedAt'] as String?,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
      vodId: json['vodId'] as String?,
      username: json['username'] as String?,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
    );
  }
}
