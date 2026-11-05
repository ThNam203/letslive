/// JSON bodies shaped like the real backend responses (verified by the
/// contract tests), for the home widget and repository tests.
library;

Map<String, Object?> livestreamJson({
  String id = 'ls-1',
  String userId = 'u-1',
  String title = 'Live title',
  String? thumbnailUrl,
  int viewCount = 1500,
}) => {
  'id': id,
  'userId': userId,
  'title': title,
  'description': null,
  'thumbnailUrl': thumbnailUrl,
  'viewCount': viewCount,
  'visibility': 'public',
  'startedAt': DateTime.now()
      .subtract(const Duration(minutes: 5))
      .toUtc()
      .toIso8601String(),
  'endedAt': null,
  'createdAt': '2026-11-04T00:00:00Z',
  'updatedAt': '2026-11-04T00:00:00Z',
  'vodId': null,
};

Map<String, Object?> vodJson({
  String id = 'vod-1',
  String userId = 'u-1',
  String title = 'Vod title',
  int duration = 3725,
}) => {
  'id': id,
  'livestreamId': null,
  'userId': userId,
  'title': title,
  'description': null,
  'thumbnailUrl': null,
  'visibility': 'public',
  'viewCount': 12,
  'likeCount': 3,
  'duration': duration,
  'playbackUrl': null,
  'status': 'ready',
  'originalFileUrl': null,
  'createdAt': '2026-11-01T00:00:00Z',
  'updatedAt': '2026-11-01T00:00:00Z',
};

Map<String, Object?> publicUserJson({
  String id = 'u-1',
  String username = 'streamer',
  String? profilePicture,
  String? bio,
  int followerCount = 7,
  bool? isFollowing,
  Map<String, Object?>? socialMediaLinks,
}) => {
  'id': id,
  'username': username,
  'followerCount': followerCount,
  'createdAt': '2026-01-15T00:00:00Z',
  'profilePicture': profilePicture,
  'bio': bio,
  'isFollowing': ?isFollowing,
  'socialMediaLinks': ?socialMediaLinks,
};
