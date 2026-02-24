package dto

import (
	"sen1or/letslive/livestream/domains"

	"github.com/gofrs/uuid/v5"
)

type CommentUser struct {
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	DisplayName    *string   `json:"displayName,omitempty"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
}

type VODCommentWithUser struct {
	domains.VODComment
	User *CommentUser `json:"user,omitempty"`
}
