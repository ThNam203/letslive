package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) DeleteComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	comment, err := s.commentRepo.GetById(ctx, commentId)
	if err != nil {
		return err
	}

	// only the author can delete their own comment
	if comment.UserId != userId {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_FORBIDDEN,
			nil,
			nil,
			nil,
		)
	}

	// if this is a reply, soft-delete + decrement parent's reply_count atomically
	if comment.ParentId != nil {
		return s.deleteReplyWithTransaction(ctx, commentId, *comment.ParentId)
	}

	return s.commentRepo.SoftDelete(ctx, commentId)
}

func (s *VODCommentService) deleteReplyWithTransaction(ctx context.Context, commentId uuid.UUID, parentId uuid.UUID) *response.Response[any] {
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [deletecomment: %v]", txErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}
	defer tx.Rollback(ctx)

	txCommentRepo := s.commentRepo.WithTx(tx)

	if softDelErr := txCommentRepo.SoftDelete(ctx, commentId); softDelErr != nil {
		return softDelErr
	}

	if decErr := txCommentRepo.DecrementReplyCount(ctx, parentId); decErr != nil {
		return decErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [deletecomment: %v]", commitErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return nil
}
