package http

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateLivestreamRequestDTO struct {
	UserId       uuid.UUID `json:"userId" validate:"required,uuid"`
	Title        *string   `json:"title" validate:"required,gte=3,lte=256"`
	Description  *string   `json:"description,omitempty" validate:"omitempty,lte=1000"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty" validate:"omitempty,url"`
	Visibility   string    `json:"visibility" validate:"oneof=public private,required"`
}

type GetLivestreamRequestDTO struct{}

type EndLivestreamRequestDTO struct {
	PlaybackURL *string   `json:"playbackUrl,omitempty" validate:"omitempty,url"`
	EndedAt     time.Time `json:"endedAt" validate:"omitempty"`
	Duration    int64     `json:"duration" validate:"omitempty"`
}

type LivestreamResponseDTO struct {
	Id           uuid.UUID  `json:"id"`
	UserId       uuid.UUID  `json:"userId"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ThumbnailURL string     `json:"thumbnailUrl"`
	Visibility   string     `json:"visibility"`
	StartedAt    *time.Time `json:"startedAt"`
	EndedAt      *time.Time `json:"endedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	VODId        *uuid.UUID `json:"vodId,omitempty"`
}
