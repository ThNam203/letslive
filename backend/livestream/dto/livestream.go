package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateLivestreamRequestDTO struct {
	UserId       uuid.UUID `json:"userId" validate:"required,uuid"`
	Title        *string   `json:"title" validate:""`
	Description  *string   `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty" validate:"omitempty,url"`
	Visibility   *string   `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
}

type UpdateLivestreamRequestDTO struct {
	Title        *string `json:"title,omitempty" validate:"omitempty,gte=3,lte=100"`
	Description  *string `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string `json:"thumbnailURL,omitempty" validate:"omitempty"`
	Visibility   *string `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
}

type UpdateVODRequestDTO struct {
	Title        *string `json:"title,omitempty" validate:"omitempty,gte=3,lte=100"`
	Description  *string `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string `json:"thumbnailURL,omitempty" validate:"omitempty"`
	Visibility   *string `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
}

type EndLivestreamRequestDTO struct {
	PlaybackURL *string    `json:"playbackUrl,omitempty" validate:"omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty" validate:"omitempty"`
	Duration    int64      `json:"duration,omitempty" validate:"omitempty"`
}
