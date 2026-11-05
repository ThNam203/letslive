/// Every input limit the app enforces, in one place. Each value is copied
/// from the web app (and the backend validators the web constants mirror) so
/// mobile accepts and rejects exactly what web does; the comment on each
/// names its web source.
library;

/// Field and upload limits shared with web.
abstract final class FieldLimits {
  /// `EMAIL_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int emailMax = 320;

  /// `PASSWORD_MIN_LENGTH` in `web/constant/password.ts`.
  static const int passwordMin = 8;

  /// `PASSWORD_MAX_LENGTH` in `web/constant/password.ts`.
  static const int passwordMax = 72;

  /// `USERNAME_MIN_LENGTH` in `web/constant/field-limits.ts`.
  static const int usernameMin = 6;

  /// `USERNAME_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int usernameMax = 30;

  /// `BIO_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int bioMax = 300;

  /// `SOCIAL_URL_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int socialUrlMax = 2048;

  /// `SEARCH_QUERY_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int searchQueryMax = 100;

  /// `CHAT_MESSAGE_MAX_LENGTH` in `web/constant/field-limits.ts`, in UTF-16
  /// units (what the server counts).
  static const int chatMessageMax = 500;

  /// `VOD_COMMENT_MAX_LENGTH` in `web/constant/field-limits.ts`, in UTF-16
  /// units.
  static const int vodCommentMax = 2000;

  /// `AVATAR_MAX_FILE_MB` in `web/constant/image.ts`.
  static const int avatarMaxFileMb = 10;

  /// `BACKGROUND_MAX_FILE_MB` in `web/constant/image.ts`.
  static const int backgroundMaxFileMb = 10;

  /// `AVATAR_MIN_DIMENSION` in `web/constant/image.ts`: smallest side of a
  /// profile picture, in pixels.
  static const int avatarMinDimension = 80;

  /// `AVATAR_SOURCE_MAX_DIMENSION` in `web/constant/image.ts`: largest side
  /// of a picked profile picture, in pixels.
  static const int avatarSourceMaxDimension = 10000;

  /// `AVATAR_ACCEPTED_TYPES` in `web/constant/image.ts`: what web's picker
  /// offers. Web re-encodes the cropped square to JPEG before uploading,
  /// which mobile does not do, so mobile checks [avatarUploadTypes].
  static const List<String> avatarAcceptedTypes = [
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
  ];

  /// The formats the server decodes for a profile picture (`avatarFormats`
  /// in `backend/user/services/avatar_image.go`); anything else answers
  /// `res_err_invalid_image` ("use JPEG or WebP").
  static const List<String> avatarUploadTypes = ['image/jpeg', 'image/webp'];
}
