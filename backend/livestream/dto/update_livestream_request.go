package dto

type UpdateLivestreamRequestDTO struct {
	Title        *string `json:"title,omitempty" validate:"omitempty,gte=3,lte=100"`
	Description  *string `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string `json:"thumbnailUrl,omitempty" validate:"omitempty"`
	Visibility   *string `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
}

