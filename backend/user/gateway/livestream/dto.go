package livestream

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type GetLivestreamResponseDTO struct {
	Id           uuid.UUID  `json:"id" db:"id"`
	UserId       uuid.UUID  `json:"userId" db:"user_id"`
	Title        *string    `json:"title" db:"title"`
	Description  *string    `json:"description" db:"description"`
	ThumbnailURL *string    `json:"thumbnailUrl" db:"thumbnail_url"`
	Status       string     `json:"status" db:"status"`
	ViewCount    int64      `json:"viewCount" db:"view_count"`
	StartedAt    *time.Time `json:"startedAt" db:"started_at"`
	EndedAt      *time.Time `json:"endedAt" db:"ended_at"`
	PlaybackURL  *string    `json:"playbackUrl" db:"playback_url"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	Duration     *int64     `json:"duration"`
}
