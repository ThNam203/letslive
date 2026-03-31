package user

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type UserPublicInfo struct {
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	DisplayName    *string   `json:"displayName,omitempty"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
}

type UserGateway interface {
	GetUserPublicInfo(ctx context.Context, userId uuid.UUID) (*UserPublicInfo, error)
}
