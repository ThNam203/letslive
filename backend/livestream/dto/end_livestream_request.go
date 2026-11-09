package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type EndLivestreamRequestDTO struct {
	PlaybackURL *string    `json:"playbackUrl,omitempty" validate:"omitempty,lte=2048"`
	EndedAt     *time.Time `json:"endedAt,omitempty" validate:"omitempty"`
	Duration    int64      `json:"duration,omitempty" validate:"omitempty"`
	// VODStatus is the status the VOD is created with: "processing" while the
	// recording is still being written, empty or "ready" otherwise.
	VODStatus string `json:"vodStatus,omitempty" validate:"omitempty,oneof=processing ready"`
}

type EndLivestreamResponseDTO struct {
	VODId *uuid.UUID `json:"vodId,omitempty"`
}
