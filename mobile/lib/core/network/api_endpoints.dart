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

  /// Custom chat commands: `GET ?roomId=<streamerId>` for a room, `POST` to
  /// create; the settings screens also use `/chat-commands/mine` and `/<id>`
  /// (PATCH, DELETE).
  static const String chatCommands = '/chat-commands';

  /// Prefix of the video service: `/vods/<id>` (detail, and PATCH/DELETE for
  /// the owner), `/vods/<id>/view`, `/vods/<id>/reaction`,
  /// `/vods/<id>/comments`, and `?userId=` for the videos of one user.
  static const String vods = '/vods';

  /// `GET ?page=&limit=` (max 20): the signed-in user's own videos, any
  /// status and visibility.
  static const String myVods = '/vods/author';

  /// `POST` multipart `file`, `title`, `description`, `visibility`: uploads a
  /// video; answers the new VOD with status `processing`.
  static const String uploadVod = '/vods/upload';

  /// `POST` multipart `file`: stores one file and answers its path (web
  /// `UploadFile`).
  static const String uploadFile = '/upload-file';

  /// `PATCH` multipart `title`, `description`, `thumbnail`, `thumbnailUrl`:
  /// the signed-in user's next-livestream information.
  static const String livestreamInformation = '/user/me/livestream-information';

  /// `PATCH`: generates a new stream key and answers it.
  static const String streamKey = '/user/me/api-key';

  /// Prefix of comment actions: `/vod-comments/<id>` (edit, delete),
  /// `/<id>/replies`, `/<id>/like` and `/liked-ids`.
  static const String vodComments = '/vod-comments';

  /// Direct messages: `/conversations` (list, create), `/<id>` (details,
  /// rename, leave), `/<id>/participants`, `/<id>/messages`, `/<id>/read`,
  /// `/<id>/typing` and `/unread-counts`.
  static const String conversations = '/conversations';

  /// Push registration of this device: `POST {token, platform, locale}`,
  /// `DELETE /<token>` (URL-encoded: FCM tokens contain a colon).
  static const String userDevices = '/user/me/devices';

  /// The user's notifications (20 per `?page=`, 0-based): `GET`,
  /// `GET /unread-count`, `PATCH /<id>/read`, `PATCH /read-all` and
  /// `DELETE /<id>`.
  static const String notifications = '/user/me/notifications';

  /// Prefix of a public profile: `/user/<id>`.
  static const String user = '/user';

  /// `GET`: the signed-in user's balances (web `GetMyWallet`).
  static const String wallet = '/wallet';

  /// `GET`: the platform's currencies (web `GetCurrencies`).
  static const String currencies = '/currencies';

  /// `GET ?page=&limit=` (max 20, 0-based page): the user's transactions.
  static const String transactions = '/transactions';

  /// `GET`: the shop catalogue (public).
  static const String shopItems = '/shop/items';

  /// `POST {shopItemId, quantity, recipientUserId?}`: buys an item, or sends
  /// it as a gift when a recipient is given. Not idempotent.
  static const String shopPurchase = '/shop/purchase';

  /// `GET ?page=&limit=`: the items the signed-in user owns.
  static const String inventory = '/user/me/inventory';

  /// `GET ?page=&limit=`: gifts the signed-in user sent. Received gifts of a
  /// user are `/user/<id>/gifts/received`.
  static const String giftsSent = '/user/me/gifts/sent';
}
