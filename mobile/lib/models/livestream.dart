class Livestream {
  final String id;
  final String userId;
  final String? title;
  final String? description;
  final String? thumbnailUrl;
  final String status;
  final String startedAt;
  final String? endedAt;
  final int viewerCount;

  // Joined user info
  final String? username;
  final String? displayName;
  final String? profilePicture;

  const Livestream({
    required this.id,
    required this.userId,
    this.title,
    this.description,
    this.thumbnailUrl,
    required this.status,
    required this.startedAt,
    this.endedAt,
    this.viewerCount = 0,
    this.username,
    this.displayName,
    this.profilePicture,
  });

  bool get isLive => status == 'started';

  factory Livestream.fromJson(Map<String, dynamic> json) {
    return Livestream(
      id: json['id'] as String,
      userId: json['userId'] as String,
      title: json['title'] as String?,
      description: json['description'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      status: json['status'] as String,
      startedAt: json['startedAt'] as String,
      endedAt: json['endedAt'] as String?,
      viewerCount: json['viewerCount'] as int? ?? 0,
      username: json['username'] as String?,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
    );
  }
}
