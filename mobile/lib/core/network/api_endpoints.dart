/// Gateway paths used by the app. Paths exclude `/v1`; Envoy rewrites `^/` to
/// `/v1/`. Only endpoints verified against `configs/envoy/envoy.yaml`, the
/// backend handlers and the web clients are listed. Keep one
/// `static const String name = '/path';` per line (parsed by parity_check).
abstract final class ApiEndpoints {
  static const String login = '/auth/login';
  static const String signup = '/auth/signup';
  static const String logout = '/auth/logout';
  static const String refreshToken = '/auth/refresh-token';
  static const String verifyEmail = '/auth/verify-email';
  static const String googleMobile = '/auth/google/mobile';
  static const String userMe = '/user/me';
  static const String following = '/user/me/following';
  static const String profilePicture = '/user/me/profile-picture';
  static const String backgroundPicture = '/user/me/background-picture';
  static const String changePassword = '/auth/password';
  static const String recommendedChannels = '/users/recommendations';
  static const String popularLivestreams = '/popular-livestreams';
  static const String popularVods = '/popular-vods';
  static const String searchUsers = '/users/search';
  static const String searchLivestreams = '/livestreams/search';
  static const String searchVods = '/vods/search';

  /// `GET ?userId=<id>`: the user's live stream, or no data when offline.
  static const String livestreams = '/livestreams';

  /// HLS host: `/transcode/<livestreamId>/index.m3u8` (Envoy rewrites it to
  /// the transcoder's `/static`, without `/v1`). Not called through the
  /// client; the player loads it.
  static const String transcode = '/transcode';

  /// Livestream chat: `GET ?roomId=` history, `POST {roomId, text}` send.
  /// Not direct messages (those are `/conversations`).
  static const String messages = '/messages';

  /// Custom chat commands: `GET ?roomId=<streamerId>` for a room; the
  /// settings screen also uses `/chat-commands/mine` and `/<id>`.
  static const String chatCommands = '/chat-commands';

  /// Prefix of the video service: `/vods/<id>` (detail), `/vods/<id>/view`,
  /// `/vods/<id>/reaction`, `/vods/<id>/comments`, and `?userId=` for the
  /// videos of one user. As a prefix it also covers `/vods/upload`,
  /// `/vods/author` and PATCH/DELETE `/vods/<id>`, which the app does not
  /// call yet (Task 19); PARITY.md lists them as planned.
  static const String vods = '/vods';

  /// Prefix of comment actions: `/vod-comments/<id>` (edit, delete),
  /// `/<id>/replies`, `/<id>/like` and `/liked-ids`.
  static const String vodComments = '/vod-comments';

  /// Direct messages: `/conversations` (list, create), `/<id>` (details,
  /// rename, leave), `/<id>/participants`, `/<id>/messages`, `/<id>/read`,
  /// `/<id>/typing` and `/unread-counts`.
  static const String conversations = '/conversations';

  /// Prefix of a public profile: `/user/<id>`.
  static const String user = '/user';
}
