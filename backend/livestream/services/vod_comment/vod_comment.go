package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	usergateway "sen1or/letslive/livestream/gateway/user"
	"sen1or/letslive/livestream/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VODCommentService struct {
	commentRepo     domains.VODCommentRepository
	commentLikeRepo domains.VODCommentLikeRepository
	vodRepo         domains.VODRepository
	userGateway     usergateway.UserGateway
	dbPool          *pgxpool.Pool
}

func NewVODCommentService(
	commentRepo domains.VODCommentRepository,
	commentLikeRepo domains.VODCommentLikeRepository,
	vodRepo domains.VODRepository,
	userGateway usergateway.UserGateway,
	dbPool *pgxpool.Pool,
) *VODCommentService {
	return &VODCommentService{
		commentRepo:     commentRepo,
		commentLikeRepo: commentLikeRepo,
		vodRepo:         vodRepo,
		userGateway:     userGateway,
		dbPool:          dbPool,
	}
}

// enrichCommentsWithUsers fetches user info for all unique user IDs in the comments
// and returns enriched DTOs. This is best-effort: if user info can't be fetched,
// the comment is returned without user data.
func (s *VODCommentService) enrichCommentsWithUsers(ctx context.Context, comments []domains.VODComment) []dto.VODCommentWithUser {
	// collect unique user IDs
	userIdSet := make(map[uuid.UUID]struct{})
	for _, c := range comments {
		if !c.IsDeleted {
			userIdSet[c.UserId] = struct{}{}
		}
	}

	// fetch user info for each unique user
	userMap := make(map[uuid.UUID]*dto.CommentUser)
	for userId := range userIdSet {
		info, err := s.userGateway.GetUserPublicInfo(ctx, userId)
		if err != nil {
			logger.Warnf(ctx, "failed to fetch user info for userId %s: %v", userId, err)
			continue // best-effort: skip if user info unavailable
		}
		if info != nil {
			userMap[userId] = &dto.CommentUser{
				Id:             info.Id,
				Username:       info.Username,
				DisplayName:    info.DisplayName,
				ProfilePicture: info.ProfilePicture,
			}
		}
	}

	// build enriched DTOs, stripping content from deleted comments
	result := make([]dto.VODCommentWithUser, len(comments))
	for i, c := range comments {
		if c.IsDeleted {
			c.Content = ""
		}
		result[i] = dto.VODCommentWithUser{
			VODComment: c,
			User:       userMap[c.UserId],
		}
	}

	return result
}
