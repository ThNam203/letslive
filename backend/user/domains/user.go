package domains

import (
	"context"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"
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
	UserStatusDisabled UserStatus = "disabled"
)

type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGoogle AuthProvider = "google"
)

type UserRepository interface {
	GetById(ctx context.Context, userId uuid.UUID) (*User, *response.Response[any])
	GetAll(ctx context.Context, page int) ([]User, *response.Response[any])
	GetByUsername(ctx context.Context, username string) (*User, *response.Response[any])
	GetByEmail(ctx context.Context, email string) (*User, *response.Response[any])
	GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*User, *response.Response[any])

	// the authenticatedUserId is used for checking if the caller is following the userId
	GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any])
	SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any])

	Create(ctx context.Context, username, email string, authProvider AuthProvider) (*User, *response.Response[any])
	Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*User, *response.Response[any])
	UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any]
	UpdateProfilePicture(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) *response.Response[any]
	UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) *response.Response[any]
}
