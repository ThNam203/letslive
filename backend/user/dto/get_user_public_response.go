package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

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

