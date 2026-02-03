package dto

type SocialMediaLinks struct {
	Facebook  *string `json:"facebook,omitempty" validate:"omitempty,url"`
	Twitter   *string `json:"twitter,omitempty" validate:"omitempty,url"`
	Instagram *string `json:"instagram,omitempty" validate:"omitempty,url"`
	LinkedIn  *string `json:"linkedin,omitempty" validate:"omitempty,url"`
	Github    *string `json:"github,omitempty" validate:"omitempty,url"`
	Youtube   *string `json:"youtube,omitempty" validate:"omitempty,url"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url"`
	TikTok    *string `json:"tiktok,omitempty" validate:"omitempty,url"`
}

