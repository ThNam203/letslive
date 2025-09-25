package domains

import (
	"context"
	"sen1or/letslive/livestream/response"
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
	GetById(ctx context.Context, id uuid.UUID) (*Livestream, *response.Response[any])
	GetByUser(ctx context.Context, userId uuid.UUID) (*Livestream, *response.Response[any])
	GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]Livestream, *response.Response[any])
	Create(ctx context.Context, ls Livestream) (*Livestream, *response.Response[any])
	Update(ctx context.Context, ls Livestream) (*Livestream, *response.Response[any])
	Delete(ctx context.Context, id uuid.UUID) *response.Response[any]
}
