package domains

import (
	"context"
	"sen1or/letslive/user/dto"
	servererrors "sen1or/letslive/user/errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	Id                uuid.UUID    `json:"id" db:"id"`
	Username          string       `json:"username" db:"username"`
	Email             string       `json:"email" db:"email"`
	Status            UserStatus   `json:"status" db:"status"`
	AuthProvider      AuthProvider `json:"authProvider" db:"auth_provider"`
	CreatedAt         time.Time    `json:"createdAt" db:"created_at"`
	StreamAPIKey      uuid.UUID    `json:"streamAPIKey" db:"stream_api_key"`
	DisplayName       *string      `json:"displayName,omitempty" db:"display_name"`
	PhoneNumber       *string      `json:"phoneNumber,omitempty" db:"phone_number"`
	Bio               *string      `json:"bio,omitempty" db:"bio"`
	ProfilePicture    *string      `json:"profilePicture,omitempty" db:"profile_picture"`
	BackgroundPicture *string      `json:"backgroundPicture,omitempty" db:"background_picture"`

	LivestreamInformation `json:"livestreamInformation"`
}

type UserStatus string

const (
	UserStatusNormal   UserStatus = "normal"
	UserStatusDisabled            = "disabled"
)

type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGoogle              = "google"
)

type UserRepository interface {
	GetById(ctx context.Context, userId uuid.UUID) (*User, *servererrors.ServerError)
	GetAll(ctx context.Context, page int) ([]User, *servererrors.ServerError)
	GetByUsername(ctx context.Context, username string) (*User, *servererrors.ServerError)
	GetByEmail(ctx context.Context, email string) (*User, *servererrors.ServerError)
	GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*User, *servererrors.ServerError)

	// the authenticatedUserId is used for checking if the caller is following the userId
	GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *servererrors.ServerError)
	SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *servererrors.ServerError)

	Create(ctx context.Context, username, email string, authProvider AuthProvider) (*User, *servererrors.ServerError)
	Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*User, *servererrors.ServerError)
	UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *servererrors.ServerError
	UpdateProfilePicture(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) *servererrors.ServerError
	UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) *servererrors.ServerError
}
