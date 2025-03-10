package dto

import (
	"github.com/gofrs/uuid/v5"
)

type GetUserResponseDTO struct {
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	IsVerified     bool      `json:"isVerified"`
	DisplayName    *string   `json:"displayName,omitempty"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
	//BackgroundPicture *string    `json:"backgroundPicture,omitempty"`
}
