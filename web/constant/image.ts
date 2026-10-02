import type { CropImageSpec, SourceImageSpec } from "@/utils/image-crop";

// mirror the *MaxFileBytes constants in
// backend/user/handlers/utils/upload_limits.go
export const AVATAR_MAX_FILE_MB = 10;
export const BACKGROUND_MAX_FILE_MB = 10;
export const LIVESTREAM_THUMBNAIL_MAX_FILE_MB = 10;
export const GENERAL_UPLOAD_MAX_FILE_MB = 10;

export const IMAGE_INPUT_ACCEPT = "image/jpeg,image/png,image/gif,image/webp";

// a decoded 10000x10000 photo already takes about 400 MB in the browser
const BROWSER_SOURCE_MAX_DIMENSION = 10000;
const UPLOAD_QUALITY = 0.92;

// The limits below mirror the rules in backend/user/services/upload_image.go.

// mirrors avatarRule
export const AVATAR_IMAGE = {
    aspect: 1,
    minCropWidth: 80,
    maxZoom: 10,
    sourceMaxDimension: BROWSER_SOURCE_MAX_DIMENSION,
    quality: UPLOAD_QUALITY,
} satisfies CropImageSpec;

// mirrors backgroundRule
export const BACKGROUND_IMAGE = {
    sourceMaxDimension: BROWSER_SOURCE_MAX_DIMENSION,
    quality: UPLOAD_QUALITY,
} satisfies SourceImageSpec;

// mirrors livestreamThumbnailRule
export const LIVESTREAM_THUMBNAIL_IMAGE = {
    aspect: 16 / 9,
    minCropWidth: 320,
    maxZoom: 10,
    sourceMaxDimension: BROWSER_SOURCE_MAX_DIMENSION,
    quality: UPLOAD_QUALITY,
} satisfies CropImageSpec;

// goes through /upload-file, so it must stay within generalImageRule
export const VOD_THUMBNAIL_IMAGE = {
    aspect: 16 / 9,
    minCropWidth: 320,
    maxZoom: 10,
    sourceMaxDimension: BROWSER_SOURCE_MAX_DIMENSION,
    quality: UPLOAD_QUALITY,
} satisfies CropImageSpec;

// mirrors generalImageRule
export const CHAT_ATTACHMENT_IMAGE = {
    sourceMaxDimension: BROWSER_SOURCE_MAX_DIMENSION,
    quality: UPLOAD_QUALITY,
} satisfies SourceImageSpec;
