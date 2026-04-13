package events

import (
	"github.com/gofrs/uuid/v5"
)

// Transcode event types.
const (
	TranscodeStreamConnected    = "transcode.stream_connected"
	TranscodeStreamDisconnected = "transcode.stream_disconnected"
	TranscodeSegmentUploaded    = "transcode.segment_uploaded"
)

// TranscodeStreamConnectedEvent is emitted when an RTMP stream connects.
type TranscodeStreamConnectedEvent struct {
	UserId      uuid.UUID `json:"userId"`
	StreamKey   string    `json:"streamKey"`
	PublishName string    `json:"publishName"`
}

// TranscodeStreamDisconnectedEvent is emitted when an RTMP stream disconnects.
type TranscodeStreamDisconnectedEvent struct {
	UserId      uuid.UUID `json:"userId"`
	PublishName string    `json:"publishName"`
}

// TranscodeSegmentUploadedEvent is emitted when an HLS segment is uploaded to storage.
type TranscodeSegmentUploadedEvent struct {
	PublishName  string `json:"publishName"`
	VariantIndex int    `json:"variantIndex"`
	RemoteID     string `json:"remoteId"`
}
