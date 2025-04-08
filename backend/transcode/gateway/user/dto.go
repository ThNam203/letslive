package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type GetUserResponseDTO struct {
	Id                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	CreatedAt         time.Time `json:"createdAt"`
	StreamAPIKey      uuid.UUID `json:"streamAPIKey"`
	DisplayName       *string   `json:"displayName,omitempty"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	ProfilePicture    *string   `json:"profilePicture,omitempty"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty"`

	LivestreamInformationResponseDTO `json:"livestreamInformation"`
}

type LivestreamInformationResponseDTO struct {
	UserID       uuid.UUID `json:"userId"`
	Title        *string   `json:"title"`
	Description  *string   `json:"description"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
}
