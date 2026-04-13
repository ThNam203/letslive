package events

import (
	"github.com/gofrs/uuid/v5"
)

// VOD event types.
const (
	VODCreated         = "vod.created"
	VODReady           = "vod.ready"
	VODTranscodeFailed = "vod.transcode_failed"
)

// VODCreatedEvent is emitted when a new VOD is created (from upload or stream-to-VOD).
type VODCreatedEvent struct {
	VODId        uuid.UUID  `json:"vodId"`
	UserId       uuid.UUID  `json:"userId"`
	Title        string     `json:"title"`
	LivestreamId *uuid.UUID `json:"livestreamId,omitempty"`
}

// VODReadyEvent is emitted when a VOD finishes transcoding and is ready for playback.
type VODReadyEvent struct {
	VODId       uuid.UUID `json:"vodId"`
	UserId      uuid.UUID `json:"userId"`
	PlaybackURL string    `json:"playbackUrl"`
	Duration    int64     `json:"duration"`
}

// VODTranscodeFailedEvent is emitted when VOD transcoding fails.
type VODTranscodeFailedEvent struct {
	VODId    uuid.UUID `json:"vodId"`
	UserId   uuid.UUID `json:"userId"`
	ErrorMsg string    `json:"errorMsg"`
}
