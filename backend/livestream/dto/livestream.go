package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// CreateLivestreamRequestDTO is used when a user starts a new livestream
type CreateLivestreamRequestDTO struct {
	Title        string     `json:"title" validate:"required,gte=3,lte=100"`
	UserId       uuid.UUID  `json:"userId" validate:"required,uuid"`
	Description  string     `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL string     `json:"thumbnailUrl,omitempty" validate:"omitempty,url"`
	Status       string     `json:"status" validate:"omitempty,lte=20"`
	StartedAt    *time.Time `json:"startedAt" validate:"omitempty,datetime"`
}

// GetLivestreamRequestDTO is used for retrieving a livestream (optional filters can be added)
type GetLivestreamRequestDTO struct{}

// UpdateLivestreamRequestDTO is used to modify an existing livestream
type UpdateLivestreamRequestDTO struct {
	Id           uuid.UUID  `json:"id" validate:"uuid"`
	Title        *string    `json:"title,omitempty" validate:"omitempty,gte=3,lte=100"`
	Description  *string    `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string    `json:"thumbnailUrl,omitempty" validate:"omitempty,url"`
	Status       *string    `json:"status,omitempty" validate:"omitempty"`
	PlaybackURL  *string    `json:"playbackUrl,omitempty" validate:"omitempty,url"`
	ViewCount    *int64     `json:"viewCount" validate:"omitempty,gte=0"`
	EndedAt      *time.Time `json:"endedAt" validate:"omitempty"`
	StartedAt    *time.Time `json:"startedAt" validate:"omitempty"`
}
