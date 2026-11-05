/// Route paths. Auth routes mirror web `app/(auth)`.
abstract final class AppRoutes {
  static const String home = '/';
  static const String search = '/search';

  /// The Notifications tab (web `/notifications`). Also where a push tap
  /// lands when it names nothing the app can open.
  static const String notifications = '/notifications';

  /// The Messages tab: the conversation list (web `/messages`).
  static const String messages = '/messages';

  /// Start a conversation. Registered before [conversationPattern], which
  /// would otherwise read "new" as a conversation id.
  static const String newConversation = '/messages/new';

  /// Route pattern of a conversation (web `/messages/[conversationId]`);
  /// push taps (type `dm`, `actionUrl` `/messages/<id>`) land here.
  static const String conversationPattern = '/messages/:conversationId';

  /// The conversation [id].
  static String conversation(String id) =>
      '/messages/${Uri.encodeComponent(id)}';

  /// The own profile tab.
  static const String profile = '/profile';

  /// Settings hub: edit profile, security, log out (web `/settings`).
  static const String settings = '/settings';

  /// Edit the own profile (web `/settings/profile`).
  static const String settingsProfile = '/settings/profile';

  /// Account security: email and password (web `/settings/security`).
  static const String settingsSecurity = '/settings/security';

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
