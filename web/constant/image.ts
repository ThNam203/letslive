export const FILE_SIZE_LIMIT_MB_UNIT = 10; // 10MB
export const FILE_SIZE_LIMIT_BYTES_UNIT = FILE_SIZE_LIMIT_MB_UNIT * 1024 * 1024; // 1MB in bytes

// mirror minAvatarDimension, maxAvatarDimension and avatarSize in
// backend/user/services/avatar_image.go
export const AVATAR_MIN_DIMENSION = 80;
export const AVATAR_MAX_DIMENSION = 10000;
export const AVATAR_SERVED_SIZE = 80;

export const AVATAR_MAX_ZOOM = 10;
// iOS Safari refuses canvases above 16.7 MP
export const AVATAR_UPLOAD_MAX_SIZE = 2048;
export const AVATAR_UPLOAD_QUALITY = 0.92;
export const AVATAR_ACCEPTED_TYPES =
    "image/jpeg,image/png,image/gif,image/webp";
