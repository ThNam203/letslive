package domains

import (
	"context"
	"sen1or/letslive/user/dto"
	"time"

	"github.com/gofrs/uuid/v5"
)

type SocialMediaLinks struct {
	Facebook  *string `json:"facebook,omitempty" validate:"omitempty,url"`
	Twitter   *string `json:"twitter,omitempty" validate:"omitempty,url"`
	Instagram *string `json:"instagram,omitempty" validate:"omitempty,url"`
	LinkedIn  *string `json:"linkedin,omitempty" validate:"omitempty,url"`
	Github    *string `json:"github,omitempty" validate:"omitempty,url"`
	Youtube   *string `json:"youtube,omitempty" validate:"omitempty,url"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url"`
	TikTok    *string `json:"tiktok,omitempty" validate:"omitempty,url"`
}

type User struct {
	Id                    uuid.UUID    `json:"id" db:"id"`
	Username              string       `json:"username" db:"username"`
	Email                 string       `json:"email" db:"email"`
	Status                UserStatus   `json:"status" db:"status"`
	AuthProvider          AuthProvider `json:"authProvider" db:"auth_provider"`
	CreatedAt             time.Time    `json:"createdAt" db:"created_at"`
	StreamAPIKey          uuid.UUID    `json:"streamAPIKey" db:"stream_api_key"`
	PhoneNumber           *string      `json:"phoneNumber,omitempty" db:"phone_number"`
	Bio                   *string      `json:"bio,omitempty" db:"bio"`
	ProfilePicture        *string      `json:"profilePicture,omitempty" db:"profile_picture"`
	BackgroundPicture     *string      `json:"backgroundPicture,omitempty" db:"background_picture"`
	FollowerCount         int          `json:"followerCount" db:"follower_count"`
	Locale                *string      `json:"locale,omitempty" db:"locale"`
	LivestreamInformation `json:"livestreamInformation,omitempty"`
	SocialMediaLinks      `json:"socialMediaLinks,omitempty"`
	SocialLinksJSON       string `json:"-"` // TODO: should i use this
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
	GetById(ctx context.Context, userId uuid.UUID) (*User, error)
	GetAll(ctx context.Context, page int) ([]User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*User, error)
	GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]UserStatus, error)
	GetDisabledUserIds(ctx context.Context) ([]uuid.UUID, error)

	// the authenticatedUserId is used for checking if the caller is following the userId
	GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, error)
	GetPublicInfosByIds(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error)
	GetRecommendedPublic(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, error)
	SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error)

	Create(ctx context.Context, username string, email string, authProvider AuthProvider) (*User, error)
	Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*User, error)
	UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) error
	UpdateProfilePicture(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) error
	UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) error
}
