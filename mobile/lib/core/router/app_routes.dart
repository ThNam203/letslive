/// Route paths. Auth routes mirror web `app/(auth)`.
abstract final class AppRoutes {
  static const String home = '/';
  static const String search = '/search';

  /// Route pattern of a channel (web `/users/[userId]`); see [channel].
  static const String channelPattern = '/users/:userId';

  /// The channel of [userId]: its live stream, or an offline notice.
  static String channel(String userId) =>
      '/users/${Uri.encodeComponent(userId)}';

  /// Route pattern of a video (web `/users/[userId]/vods/[vodId]`); see
  /// [vod].
  static const String vodPattern = '/users/:userId/vods/:vodId';

  /// The video [vodId] of [userId], with its comments.
  static String vod(String userId, String vodId) =>
      '/users/${Uri.encodeComponent(userId)}'
      '/vods/${Uri.encodeComponent(vodId)}';
  static const String login = '/login';
  static const String signup = '/signup';
  static const String accountSetup = '/account-setup';

  /// Shown while the session loads, or with a retry if loading failed.
  static const String splash = '/splash';
}
