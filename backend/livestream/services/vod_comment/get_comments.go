package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) GetCommentsByVODId(ctx context.Context, vodId uuid.UUID, page int, limit int) ([]dto.VODCommentWithUser, int, *response.Response[any]) {
	comments, err := s.commentRepo.GetByVODId(ctx, vodId, page, limit)
	if err != nil {
		return nil, 0, err
	}

	total, countErr := s.commentRepo.CountByVODId(ctx, vodId)
	if countErr != nil {
		return nil, 0, countErr
	}

	return s.enrichCommentsWithUsers(ctx, comments), total, nil
}
