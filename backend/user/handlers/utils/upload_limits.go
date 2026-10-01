package utils

// mirror the *_MAX_FILE_MB constants in web/constant/image.ts, which check
// file.size before uploading
const (
	AvatarMaxFileBytes              = 10 * 1024 * 1024
	BackgroundMaxFileBytes          = 10 * 1024 * 1024
	LivestreamThumbnailMaxFileBytes = 10 * 1024 * 1024
	GeneralUploadMaxFileBytes       = 10 * 1024 * 1024
)
