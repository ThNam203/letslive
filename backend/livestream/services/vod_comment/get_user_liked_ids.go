package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) GetUserLikedCommentIds(ctx context.Context, data dto.GetUserLikedCommentIdsRequestDTO, userId uuid.UUID) ([]uuid.UUID, *response.Response[any]) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, response.NewResponseWithValidationErrors[any](nil, nil, err)
	}

	commentUUIDs := make([]uuid.UUID, 0, len(data.CommentIds))
	for _, idStr := range data.CommentIds {
		id, err := uuid.FromString(idStr)
		if err != nil {
			return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil)
		}
		commentUUIDs = append(commentUUIDs, id)
	}

	return s.commentLikeRepo.GetUserLikedCommentIds(ctx, commentUUIDs, userId)
}
