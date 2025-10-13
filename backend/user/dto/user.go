package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateUserRequestDTO struct {
	Username     string `json:"username" validate:"required,gte=4,lte=50"`
	Email        string `json:"email" validate:"required,email"`
	AuthProvider string `json:"authProvider" validate:"oneof=google local"`
}

type GetUserRequestDTO struct{}

type GetUserPublicResponseDTO struct {
	Id                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	CreatedAt         time.Time `json:"createdAt"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	DisplayName       *string   `json:"displayName,omitempty"`
	ProfilePicture    *string   `json:"profilePicture,omitempty"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty"`
	FollowerCount     *int      `json:"followerCount,omitempty"`

	// whether or not the current fetching user is following the fetched user
	IsFollowing *bool `json:"isFollowing,omitempty"`

	LivestreamInformation `json:"livestreamInformation"`
	SocialMediaLinks      *SocialMediaLinks `json:"socialMediaLinks,omitempty"`
	SocialLinksJSON       string            `json:"-"`
}

type LivestreamInformation struct {
	Title        *string `db:"title,omitempty" json:"title"`
	Description  *string `db:"description,omitempty" json:"description"`
	ThumbnailURL *string `db:"thumbnail_url,omitempty" json:"thumbnailUrl"`
}

type GetUserByStreamAPIKeyRequestDTO struct{}

type SocialMediaLinks struct {
	Facebook  *string `json:"facebook,omitempty" validate:"omitempty,url"`
	Twitter   *string `json:"twitter,omitempty" validate:"omitempty,url"`
	Instagram *string `json:"instagram,omitempty" validate:"omitempty,url"`
	LinkedIn  *string `json:"linkedin,omitempty" validate:"omitempty,url"`
	Github    *string `json:"github,omitempty" validate:"omitempty,url"`
	Youtube   *string `json:"youtube,omitempty" validate:"omitempty,url"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url"`
}

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
