package domains

import (
	"context"
	servererrors "sen1or/letslive/user/errors"

	"github.com/gofrs/uuid/v5"
)

type LivestreamInformation struct {
	UserID       uuid.UUID `db:"user_id,omitempty" json:"userId"`
	Title        *string   `db:"title,omitempty" json:"title"`
	Description  *string   `db:"description,omitempty" json:"description"`
	ThumbnailURL *string   `db:"thumbnail_url,omitempty" json:"thumbnailUrl"`
}

type LivestreamInformationRepository interface {
	GetByUserId(context.Context, uuid.UUID) (*LivestreamInformation, *servererrors.ServerError)
	Create(context.Context, uuid.UUID) *servererrors.ServerError
	Update(context.Context, LivestreamInformation) (*LivestreamInformation, *servererrors.ServerError)
}
