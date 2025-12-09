package dto

import (
	"github.com/gofrs/uuid/v5"
)

// TODO: remove id from dto
type UpdateUserRequestDTO struct {
	Id               uuid.UUID         `json:"id" validate:"uuid"`
	Username         *string           `json:"username,omitempty" validate:"omitempty,gte=6,lte=20"`
	Status           *string           `json:"status,omitempty" validate:"oneof=normal disabled"`
	PhoneNumber      *string           `json:"phoneNumber,omitempty"`
	Bio              *string           `json:"bio,omitempty"`
	DisplayName      *string           `json:"displayName,omitempty" validate:"omitempty,gte=6,lte=20"`
	SocialMediaLinks *SocialMediaLinks `json:"socialMediaLinks,omitempty"`
}

