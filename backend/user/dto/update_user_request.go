package dto

import (
	"github.com/gofrs/uuid/v5"
)

// TODO: remove id from dto
type UpdateUserRequestDTO struct {
	Id               uuid.UUID         `json:"id" validate:"uuid"`
	Username         *string           `json:"username,omitempty" validate:"omitempty,gte=6,lte=30"`
	Status           *string           `json:"status,omitempty" validate:"omitempty,oneof=normal disabled"`
	PhoneNumber      *string           `json:"phoneNumber,omitempty" validate:"omitempty,lte=20"`
	Bio              *string           `json:"bio,omitempty" validate:"omitempty,lte=300"`
	SocialMediaLinks *SocialMediaLinks `json:"socialMediaLinks,omitempty"`
}
