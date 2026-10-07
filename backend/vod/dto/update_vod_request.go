package dto

type UpdateVODRequestDTO struct {
	Title        *string `json:"title,omitempty" validate:"omitnil,min=1,max=255"`
	Description  *string `json:"description,omitempty" validate:"omitnil,max=1000"`
	ThumbnailURL *string `json:"thumbnailUrl,omitempty" validate:"omitempty,url,lte=2048"`
	Visibility   *string `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
}
