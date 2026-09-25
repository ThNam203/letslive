package vodcomment

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) DeleteComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error {
	comment, err := s.commentRepo.GetById(ctx, commentId)
	if err != nil {
		return err
	}

	// only the author can delete their own comment
	if comment.UserId != userId {
		return domains.ErrForbidden
	}

	// if this is a reply, soft-delete + decrement parent's reply_count atomically
	if comment.ParentId != nil {
		return s.deleteReplyWithTransaction(ctx, commentId, *comment.ParentId)
	}

	return s.commentRepo.SoftDelete(ctx, commentId)
}

func (s *VODCommentService) deleteReplyWithTransaction(ctx context.Context, commentId uuid.UUID, parentId uuid.UUID) error {
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [deletecomment: %v]", txErr)
		return domains.ErrDatabaseIssue
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
		return domains.ErrDatabaseIssue
	}

	return nil
}
