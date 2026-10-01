export const FILE_SIZE_LIMIT_MB_UNIT = 10; // 10MB
export const FILE_SIZE_LIMIT_BYTES_UNIT = FILE_SIZE_LIMIT_MB_UNIT * 1024 * 1024; // 1MB in bytes

// mirror minAvatarDimension, maxAvatarDimension and avatarSize in
// backend/user/services/avatar_image.go
export const AVATAR_MIN_DIMENSION = 80;
export const AVATAR_MAX_DIMENSION = 10000;
export const AVATAR_SERVED_SIZE = 80;

export const AVATAR_MAX_ZOOM = 10;
export const AVATAR_PREVIEW_SIZE = 256;
export const AVATAR_ACCEPTED_TYPES =
    "image/jpeg,image/png,image/gif,image/webp";
