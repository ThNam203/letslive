package domains

import (
	servererrors "sen1or/letslive/livestream/errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

type Livestream struct {
	Id           uuid.UUID            `json:"id" db:"id"`
	UserId       uuid.UUID            `json:"userId" db:"user_id"`
	Title        *string              `json:"title" db:"title"`
	Description  *string              `json:"description" db:"description"`
	ThumbnailURL *string              `json:"thumbnailUrl" db:"thumbnail_url"`
	Status       string               `json:"status" db:"status"`
	Visibility   LivestreamVisibility `json:"visibility" db:"visibility"`
	ViewCount    int64                `json:"viewCount" db:"view_count"`
	StartedAt    *time.Time           `json:"startedAt" db:"started_at"`
	EndedAt      *time.Time           `json:"endedAt" db:"ended_at"`
	PlaybackURL  *string              `json:"playbackUrl" db:"playback_url"`
	CreatedAt    time.Time            `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time            `json:"updatedAt" db:"updated_at"`
	Duration     *int64               `json:"duration" db:"duration"`
}

type LivestreamVisibility string

const (
	LivestreamPublicVisibility  LivestreamVisibility = "public"
	LivestreamPrivateVisibility                      = "private"
)

type LivestreamRepository interface {
	GetById(uuid.UUID) (*Livestream, *servererrors.ServerError)
	GetByUser(uuid.UUID) ([]Livestream, *servererrors.ServerError)

	GetAllLivestreamings(page int) ([]Livestream, *servererrors.ServerError)
	GetPopularVODs(page int) ([]Livestream, *servererrors.ServerError)
	AddOneToViewCount(uuid.UUID)

	CheckIsUserLivestreaming(uuid.UUID) (bool, *servererrors.ServerError)

	Create(Livestream) (*Livestream, *servererrors.ServerError)
	Update(Livestream) (*Livestream, *servererrors.ServerError)
	Delete(uuid.UUID) *servererrors.ServerError
}
