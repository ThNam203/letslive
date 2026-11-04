package dto

const MaxDeviceTokenLength = 4096

type RegisterDeviceRequestDTO struct {
	Token    string  `json:"token" validate:"required,max=4096"`
	Platform string  `json:"platform" validate:"required,oneof=ios android"`
	Locale   *string `json:"locale,omitempty" validate:"omitempty,bcp47_language_tag,max=35"`
}
