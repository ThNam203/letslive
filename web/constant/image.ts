// mirror the *MaxFileBytes constants in
// backend/user/handlers/utils/upload_limits.go
export const AVATAR_MAX_FILE_MB = 10;
export const BACKGROUND_MAX_FILE_MB = 10;
export const LIVESTREAM_THUMBNAIL_MAX_FILE_MB = 10;
export const GENERAL_UPLOAD_MAX_FILE_MB = 10;

// mirror minAvatarDimension, maxAvatarDimension and avatarSize in
// backend/user/services/avatar_image.go
export const AVATAR_MIN_DIMENSION = 80;
export const AVATAR_MAX_DIMENSION = 10000;
export const AVATAR_SERVED_DIMENSION = 80;

export const AVATAR_MAX_ZOOM = 10;
// iOS Safari refuses canvases above 16.7 MP
export const AVATAR_UPLOAD_MAX_DIMENSION = 2048;
export const AVATAR_UPLOAD_QUALITY = 0.92;
export const AVATAR_ACCEPTED_TYPES =
    "image/jpeg,image/png,image/gif,image/webp";
