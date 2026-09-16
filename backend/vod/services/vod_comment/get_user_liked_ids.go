package vodcomment

import (
	"context"
	"fmt"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/vod/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) GetUserLikedCommentIds(ctx context.Context, data dto.GetUserLikedCommentIdsRequestDTO, userId uuid.UUID) ([]uuid.UUID, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, fmt.Errorf("%w: %w", domains.ErrInvalidInput, err)
	}

	commentUUIDs := make([]uuid.UUID, 0, len(data.CommentIds))
	for _, idStr := range data.CommentIds {
		id, err := uuid.FromString(idStr)
		if err != nil {
			return nil, domains.ErrInvalidInput
		}
		commentUUIDs = append(commentUUIDs, id)
	}

	return s.commentLikeRepo.GetUserLikedCommentIds(ctx, commentUUIDs, userId)
}
