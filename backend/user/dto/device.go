package dto

// MaxDeviceTokenLength is the longest FCM token accepted (the validate tag
// below repeats it).
const MaxDeviceTokenLength = 4096

// RegisterDeviceRequestDTO is the body of POST /user/me/devices: this
// device's FCM token, its platform and, optionally, its app language.
type RegisterDeviceRequestDTO struct {
	Token    string  `json:"token" validate:"required,max=4096"`
	Platform string  `json:"platform" validate:"required,oneof=ios android"`
	Locale   *string `json:"locale,omitempty" validate:"omitempty,bcp47_language_tag,max=35"`
}
