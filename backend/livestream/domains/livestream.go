package domains

import (
	"context"
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
	GetById(ctx context.Context, id uuid.UUID) (*Livestream, error)
	GetByUser(ctx context.Context, userId uuid.UUID) (*Livestream, error)
	GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]Livestream, error)
	Create(ctx context.Context, ls Livestream) (*Livestream, error)
	Update(ctx context.Context, ls Livestream) (*Livestream, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error
	ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error
}
