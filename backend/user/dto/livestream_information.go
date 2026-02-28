package dto

type LivestreamInformation struct {
	Title        *string `db:"title,omitempty" json:"title" validate:"omitempty,lte=50"`
	Description  *string `db:"description,omitempty" json:"description" validate:"omitempty,lte=500"`
	ThumbnailURL *string `db:"thumbnail_url,omitempty" json:"thumbnailUrl" validate:"omitempty,url,lte=2048"`
}

