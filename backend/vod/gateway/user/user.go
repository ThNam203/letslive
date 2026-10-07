package user

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

// Notification types sent to the user service.
const (
	NotificationTypeVODComment      = "vod_comment"
	NotificationTypeVODCommentReply = "vod_comment_reply"
)

type UserPublicInfo struct {
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
}

type CreateNotificationRequest struct {
	UserId      string  `json:"userId"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Message     string  `json:"message"`
	ActionUrl   *string `json:"actionUrl,omitempty"`
	ActionLabel *string `json:"actionLabel,omitempty"`
	ReferenceId *string `json:"referenceId,omitempty"`
}

type UserGateway interface {
	GetUserPublicInfo(ctx context.Context, userId uuid.UUID) (*UserPublicInfo, error)
	CreateNotification(ctx context.Context, req CreateNotificationRequest) error
}
