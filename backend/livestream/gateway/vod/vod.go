package vod

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type CreateVODRequest struct {
	LivestreamId string `json:"livestreamId"`
	UserId       string `json:"userId"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	PlaybackURL  string `json:"playbackUrl,omitempty"`
	Duration     int64  `json:"duration"`
}

type VODGateway interface {
	CreateVOD(ctx context.Context, req CreateVODRequest) (*uuid.UUID, error)
}
