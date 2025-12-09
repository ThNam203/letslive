package dto

import (
	"sen1or/letslive/livestream/domains"

	"github.com/gofrs/uuid/v5"
)

type CreateLivestreamRequestDTO struct {
	UserId       uuid.UUID                     `json:"userId" validate:"required,uuid"`
	Title        *string                       `json:"title" validate:""`
	Description  *string                       `json:"description,omitempty" validate:"omitempty,lte=500"`
	ThumbnailURL *string                       `json:"thumbnailUrl,omitempty" validate:"omitempty,url"`
	Visibility   *domains.LivestreamVisibility `json:"visibility,omitempty" validate:"required,oneof=public private"`
}

