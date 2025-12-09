package dto

import "time"

type EndLivestreamRequestDTO struct {
	PlaybackURL *string    `json:"playbackUrl,omitempty" validate:"omitempty"`
	EndedAt     *time.Time `json:"endedAt,omitempty" validate:"omitempty"`
	Duration    int64      `json:"duration,omitempty" validate:"omitempty"`
}

