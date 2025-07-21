package domains

import (
	"context"
	serviceresponse "sen1or/letslive/livestream/responses"
	"time"

	"github.com/gofrs/uuid/v5"
)

type VODVisibility string

const (
	VODPublicVisibility  VODVisibility = "public"
	VODPrivateVisibility VODVisibility = "private"
)

type VOD struct {
	Id           uuid.UUID     `json:"id" db:"id"`
	LivestreamId uuid.UUID     `json:"livestreamId" db:"livestream_id"`
	UserId       uuid.UUID     `json:"userId" db:"user_id"`
	Title        string        `json:"title" db:"title"`
	Description  *string       `json:"description" db:"description"`
	ThumbnailURL *string       `json:"thumbnailUrl" db:"thumbnail_url"`
	Visibility   VODVisibility `json:"visibility" db:"visibility"`
	ViewCount    int64         `json:"viewCount" db:"view_count"`
	Duration     int64         `json:"duration" db:"duration"`
	PlaybackURL  *string       `json:"playbackUrl" db:"playback_url"`
	CreatedAt    time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time     `json:"updatedAt" db:"updated_at"`
}

type VODRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*VOD, *serviceresponse.ServiceErrorResponse)
	GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]VOD, *serviceresponse.ServiceErrorResponse)
	GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]VOD, *serviceresponse.ServiceErrorResponse)
	GetPopular(ctx context.Context, page int, limit int) ([]VOD, *serviceresponse.ServiceErrorResponse)
	IncrementViewCount(ctx context.Context, id uuid.UUID) *serviceresponse.ServiceErrorResponse
	Create(ctx context.Context, vod VOD) (*VOD, *serviceresponse.ServiceErrorResponse)
	Update(ctx context.Context, vod VOD) (*VOD, *serviceresponse.ServiceErrorResponse)
	Delete(ctx context.Context, id uuid.UUID) *serviceresponse.ServiceErrorResponse
}
