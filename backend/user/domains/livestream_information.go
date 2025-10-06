package domains

import (
	"context"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type LivestreamInformation struct {
	UserID       uuid.UUID `db:"user_id,omitempty" json:"userId"`
	Title        *string   `db:"title,omitempty" json:"title"`
	Description  *string   `db:"description,omitempty" json:"description"`
	ThumbnailURL *string   `db:"thumbnail_url,omitempty" json:"thumbnailUrl"`
}

type LivestreamInformationRepository interface {
	GetByUserId(context.Context, uuid.UUID) (*LivestreamInformation, *response.Response[any])
	Create(context.Context, uuid.UUID) *response.Response[any]
	Update(context.Context, LivestreamInformation) (*LivestreamInformation, *response.Response[any])
}
