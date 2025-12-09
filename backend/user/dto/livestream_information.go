package dto

type LivestreamInformation struct {
	Title        *string `db:"title,omitempty" json:"title"`
	Description  *string `db:"description,omitempty" json:"description"`
	ThumbnailURL *string `db:"thumbnail_url,omitempty" json:"thumbnailUrl"`
}

