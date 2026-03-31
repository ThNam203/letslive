class Vod {
  final String id;
  final String? livestreamId;
  final String userId;
  final String title;
  final String? description;
  final String? thumbnailUrl;
  final String visibility;
  final int viewCount;
  final int duration;
  final String? playbackUrl;
  final String status;
  final String? originalFileUrl;
  final String createdAt;
  final String updatedAt;

  // Joined user info (from backend joined queries)
  final String? username;
  final String? displayName;
  final String? profilePicture;

  const Vod({
    required this.id,
    this.livestreamId,
    required this.userId,
    required this.title,
    this.description,
    this.thumbnailUrl,
    this.visibility = 'public',
    this.viewCount = 0,
    this.duration = 0,
    this.playbackUrl,
    this.status = 'ready',
    this.originalFileUrl,
    required this.createdAt,
    required this.updatedAt,
    this.username,
    this.displayName,
    this.profilePicture,
  });

  bool get isPublic => visibility == 'public';
  bool get isReady => status == 'ready';
  bool get isProcessing => status == 'processing';
  bool get isFailed => status == 'failed';

  Vod copyWith({
    String? title,
    String? description,
    String? thumbnailUrl,
    String? visibility,
    int? viewCount,
    String? status,
  }) {
    return Vod(
      id: id,
      livestreamId: livestreamId,
      userId: userId,
      title: title ?? this.title,
      description: description ?? this.description,
      thumbnailUrl: thumbnailUrl ?? this.thumbnailUrl,
      visibility: visibility ?? this.visibility,
      viewCount: viewCount ?? this.viewCount,
      duration: duration,
      playbackUrl: playbackUrl,
      status: status ?? this.status,
      originalFileUrl: originalFileUrl,
      createdAt: createdAt,
      updatedAt: updatedAt,
      username: username,
      displayName: displayName,
      profilePicture: profilePicture,
    );
  }

  factory Vod.fromJson(Map<String, dynamic> json) {
    return Vod(
      id: json['id'] as String,
      livestreamId: json['livestreamId'] as String?,
      userId: json['userId'] as String,
      title: json['title'] as String,
      description: json['description'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      visibility: json['visibility'] as String? ?? 'public',
      viewCount: json['viewCount'] as int? ?? 0,
      duration: json['duration'] as int? ?? 0,
      playbackUrl: json['playbackUrl'] as String?,
      status: json['status'] as String? ?? 'ready',
      originalFileUrl: json['originalFileUrl'] as String?,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
      username: json['username'] as String?,
      displayName: json['displayName'] as String?,
      profilePicture: json['profilePicture'] as String?,
    );
  }
}
