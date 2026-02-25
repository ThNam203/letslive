package dto

import (
	"sen1or/letslive/livestream/domains"

	"github.com/gofrs/uuid/v5"
)

type CreateLivestreamRequestDTO struct {
	UserId       uuid.UUID                     `json:"userId" validate:"required,uuid"`
	Title        *string                       `json:"title" validate:"omitempty,lte=255"`
	Description  *string                       `json:"description,omitempty" validate:"omitempty,lte=1000"`
	ThumbnailURL *string                       `json:"thumbnailUrl,omitempty" validate:"omitempty,url,lte=2048"`
	Visibility   *domains.LivestreamVisibility `json:"visibility,omitempty" validate:"required,oneof=public private"`
}

