package events

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Livestream event types.
const (
	LivestreamStarted = "livestream.started"
	LivestreamEnded   = "livestream.ended"
	LivestreamUpdated = "livestream.updated"
)

// LivestreamStartedEvent is emitted when a livestream begins.
type LivestreamStartedEvent struct {
	LivestreamId uuid.UUID `json:"livestreamId"`
	UserId       uuid.UUID `json:"userId"`
	Title        string    `json:"title"`
	StartedAt    time.Time `json:"startedAt"`
}

// LivestreamEndedEvent is emitted when a livestream ends.
type LivestreamEndedEvent struct {
	LivestreamId uuid.UUID  `json:"livestreamId"`
	UserId       uuid.UUID  `json:"userId"`
	EndedAt      time.Time  `json:"endedAt"`
	Duration     int64      `json:"duration"`
	PlaybackURL  *string    `json:"playbackUrl,omitempty"`
	VODId        *uuid.UUID `json:"vodId,omitempty"`
}

// LivestreamUpdatedEvent is emitted when livestream metadata changes.
type LivestreamUpdatedEvent struct {
	LivestreamId uuid.UUID `json:"livestreamId"`
	UserId       uuid.UUID `json:"userId"`
	Title        *string   `json:"title,omitempty"`
	Description  *string   `json:"description,omitempty"`
	ThumbnailURL *string   `json:"thumbnailUrl,omitempty"`
}
