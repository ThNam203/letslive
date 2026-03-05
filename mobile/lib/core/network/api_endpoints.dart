/// All API endpoints matching the web app's API modules.
abstract final class ApiEndpoints {
  // Auth
  static const authSignup = '/auth/signup';
  static const authLogin = '/auth/login';
  static const authLogout = '/auth/logout';
  static const authRefreshToken = '/auth/refresh-token';
  static const authPassword = '/auth/password';
  static const authVerifyEmail = '/auth/verify-email';
  static const authVerifyOtp = '/auth/verify-otp';

  // User – singular "/user" for specific-user operations,
  //         plural "/users" for collection queries (matches backend).
  static const userMe = '/user/me';
  static String userById(String id) => '/user/$id';
  static const usersSearch = '/users/search';
  static const usersRecommendations = '/users/recommendations';
  static String userFollow(String id) => '/user/$id/follow';
  static String userUnfollow(String id) => '/user/$id/unfollow';
  static const userProfilePicture = '/user/me/profile-picture';
  static const userBackgroundPicture = '/user/me/background-picture';
  static const userFollowing = '/user/me/following';
  static const userLivestreamInformation = '/user/me/livestream-information';
  static const userApiKey = '/user/me/api-key';
  static String userVods(String id) => '/user/$id/vods';

  // Notifications (nested under /user/me/)
  static const notifications = '/user/me/notifications';
  static const notificationsUnreadCount =
      '/user/me/notifications/unread-count';
  static String notificationRead(String id) =>
      '/user/me/notifications/$id/read';
  static const notificationsReadAll = '/user/me/notifications/read-all';
  static String notificationById(String id) => '/user/me/notifications/$id';

  // Livestream
  static const livestreams = '/livestreams';
  static const popularLivestreams = '/popular-livestreams';
  static String livestreamTranscode(String livestreamId) =>
      '/transcode/$livestreamId/index.m3u8';

  // VOD
  static const vods = '/vods';
  static String vodById(String id) => '/vods/$id';
  static const vodsAuthor = '/vods/author';
  static const popularVods = '/popular-vods';

  // VOD Comments
  static String vodComments(String vodId) => '/vods/$vodId/comments';
  static String vodCommentById(String commentId) =>
      '/vod-comments/$commentId';
  static String vodCommentReplies(String commentId) =>
      '/vod-comments/$commentId/replies';
  static String vodCommentLike(String commentId) =>
      '/vod-comments/$commentId/like';
  static const vodCommentLikedIds = '/vod-comments/liked-ids';

  // Chat (live chat messages, roomId passed as query parameter)
  static const chatMessages = '/messages';

  // DM / Conversations (no /dm/ prefix, matches web)
  static const conversations = '/conversations';
  static const conversationsUnreadCounts = '/conversations/unread-counts';
  static String conversationById(String id) => '/conversations/$id';
  static String conversationMessages(String id) =>
      '/conversations/$id/messages';
  static String conversationParticipants(String id) =>
      '/conversations/$id/participants';
  static String conversationRemoveParticipant(String id, String userId) =>
      '/conversations/$id/participants/$userId';
  static String conversationRead(String id) => '/conversations/$id/read';

  // Upload
  static const uploadFile = '/upload-file';

  /// Paths to skip token refresh on 401.
  static const refreshExcludePaths = [
    authLogin,
    authSignup,
    authRefreshToken,
    authLogout,
  ];
}
