package user

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type UserPublicInfo struct {
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
}

type UserGateway interface {
	GetUserPublicInfo(ctx context.Context, userId uuid.UUID) (*UserPublicInfo, error)
	SearchUserIds(ctx context.Context, query string, limit int) ([]uuid.UUID, error)
}
