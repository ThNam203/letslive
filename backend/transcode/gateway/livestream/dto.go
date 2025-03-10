package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// CreateLivestreamRequestDTO is used when a user starts a new livestream
type CreateLivestreamRequestDTO struct {
	UserId       uuid.UUID `json:"userId" validate:"required,uuid"`
	Title        *string   `json:"title" validate:"required,gte=3,lte=100"`
	Description  *string   `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty" validate:"omitempty,url"`
	Status       string    `json:"status" validate:"required,lte=20"`
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
	Duration     int64      `json:"duration"`
}

type LivestreamResponseDTO struct {
	Id           uuid.UUID  `json:"id"`
	UserId       uuid.UUID  `json:"userId"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ThumbnailURL string     `json:"thumbnailUrl"`
	Status       string     `json:"status"`
	ViewCount    int64      `json:"viewCount"`
	StartedAt    *time.Time `json:"startedAt"`
	EndedAt      *time.Time `json:"endedAt"`
	PlaybackURL  string     `json:"playbackUrl"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
