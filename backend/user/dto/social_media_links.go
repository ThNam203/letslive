package dto

type SocialMediaLinks struct {
	Facebook  *string `json:"facebook,omitempty" validate:"omitempty,url,lte=2048"`
	Twitter   *string `json:"twitter,omitempty" validate:"omitempty,url,lte=2048"`
	Instagram *string `json:"instagram,omitempty" validate:"omitempty,url,lte=2048"`
	LinkedIn  *string `json:"linkedin,omitempty" validate:"omitempty,url,lte=2048"`
	Github    *string `json:"github,omitempty" validate:"omitempty,url,lte=2048"`
	Youtube   *string `json:"youtube,omitempty" validate:"omitempty,url,lte=2048"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url,lte=2048"`
	TikTok    *string `json:"tiktok,omitempty" validate:"omitempty,url,lte=2048"`
}
