package domains

import (
	"context"
	serviceresponse "sen1or/letslive/livestream/responses"
	"time"

	"github.com/gofrs/uuid/v5"
)

type LivestreamVisibility string

const (
	LivestreamPublicVisibility  LivestreamVisibility = "public"
	LivestreamPrivateVisibility                      = "private"
)

type Livestream struct {
	Id           uuid.UUID            `json:"id" db:"id"`
	UserId       uuid.UUID            `json:"userId" db:"user_id"`
	Title        string               `json:"title" db:"title"`
	Description  *string              `json:"description" db:"description"`
	ThumbnailURL *string              `json:"thumbnailUrl" db:"thumbnail_url"`
	ViewCount    int                  `json:"viewCount" db:"view_count"`
	Visibility   LivestreamVisibility `json:"visibility" db:"visibility"`
	StartedAt    time.Time            `json:"startedAt" db:"started_at"`
	EndedAt      *time.Time           `json:"endedAt" db:"ended_at"`
	CreatedAt    time.Time            `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time            `json:"updatedAt" db:"updated_at"`
	VODId        *uuid.UUID           `json:"vodId" db:"vod_id"`
}

type LivestreamRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*Livestream, *serviceresponse.ServiceErrorResponse)
	GetByUser(ctx context.Context, userId uuid.UUID) (*Livestream, *serviceresponse.ServiceErrorResponse)
	GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]Livestream, *serviceresponse.ServiceErrorResponse)
	Create(ctx context.Context, ls Livestream) (*Livestream, *serviceresponse.ServiceErrorResponse)
	Update(ctx context.Context, ls Livestream) (*Livestream, *serviceresponse.ServiceErrorResponse)
	Delete(ctx context.Context, id uuid.UUID) *serviceresponse.ServiceErrorResponse
}
