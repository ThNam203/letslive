package domains

import "github.com/gofrs/uuid/v5"

type LivestreamInformation struct {
	UserID       uuid.UUID `db:"user_id,omitempty" json:"userId"`
	Title        *string   `db:"title,omitempty" json:"title"`
	Description  *string   `db:"description,omitempty" json:"description"`
	ThumbnailURL *string   `db:"thumbnail_url,omitempty" json:"thumbnailUrl"`
}
