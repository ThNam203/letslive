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

  /// `DM_MESSAGE_MAX_LENGTH` in `web/constant/field-limits.ts` (backend
  /// `MaxDmMessageLength`), in UTF-16 units.
  static const int dmMessageMax = 2000;

  /// `GROUP_NAME_MAX_LENGTH` in `web/constant/field-limits.ts` (backend
  /// `MaxConversationNameLen`), in UTF-16 units.
  static const int groupNameMax = 100;

  /// `GROUP_MAX_MEMBERS` in `web/constant/field-limits.ts`, the creator
  /// included.
  static const int groupMaxMembers = 100;

  /// `VOD_COMMENT_MAX_LENGTH` in `web/constant/field-limits.ts`, in UTF-16
  /// units.
  static const int vodCommentMax = 2000;

  /// `STREAM_TITLE_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int streamTitleMax = 50;

  /// `STREAM_DESCRIPTION_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int streamDescriptionMax = 500;

  /// `VOD_TITLE_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int vodTitleMax = 255;

  /// `VOD_DESCRIPTION_MAX_LENGTH` in `web/constant/field-limits.ts`.
  static const int vodDescriptionMax = 1000;

  /// `MAX_FILE_SIZE` (100 MB) in `web/app/(main)/settings/upload/page.tsx`
  /// and `maxUploadSize` in `backend/vod/handlers/vod/upload_vod_private.go`.
  static const int vodUploadMaxMb = 100;

  /// `ALLOWED_EXTENSIONS` in `web/app/(main)/settings/upload/page.tsx`
  /// (also checked by the video service), lowercase with the dot.
  static const List<String> vodUploadExtensions = [
    '.mp4',
    '.mov',
    '.avi',
    '.mkv',
    '.webm',
  ];

  /// `MAX_CONCURRENT` in `web/hooks/use-upload-store.ts`: video uploads
  /// running at once; the rest wait.
  static const int vodUploadConcurrency = 3;

  /// `NAME_PATTERN` (1-32 characters) in
  /// `web/app/(main)/settings/chat-commands/page.tsx`.
  static const int chatCommandNameMax = 32;

  /// `MAX_RESPONSE` in `web/app/(main)/settings/chat-commands/page.tsx`.
  static const int chatCommandResponseMax = 500;

  /// `MAX_DESCRIPTION` in `web/app/(main)/settings/chat-commands/page.tsx`.
  static const int chatCommandDescriptionMax = 120;

  /// `LIVESTREAM_THUMBNAIL_MAX_FILE_MB` in `web/constant/image.ts`.
  static const int livestreamThumbnailMaxFileMb = 10;

  /// `GENERAL_UPLOAD_MAX_FILE_MB` in `web/constant/image.ts`: the limit of
  /// `POST /upload-file` (DM pictures, VOD thumbnails).
  static const int generalUploadMaxFileMb = 10;

  /// `MAX_FILES` in `web/app/(main)/messages/_components/message-input.tsx`:
  /// pictures per direct message.
  static const int dmImagesMax = 10;

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

  /// The formats the server decodes for a profile picture (`avatarFormats`
  /// in `backend/user/services/avatar_image.go`); anything else answers
  /// `res_err_invalid_image` ("use JPEG or WebP"). Web's picker offers more
  /// (`AVATAR_ACCEPTED_TYPES` in `web/constant/image.ts`: also PNG and GIF)
  /// because it re-encodes the cropped square to JPEG first; mobile uploads
  /// the file as is, so it checks this narrower list.
  static const List<String> avatarUploadTypes = ['image/jpeg', 'image/webp'];
}
