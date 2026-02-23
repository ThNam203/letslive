package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) LikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	// verify comment exists
	_, err := s.commentRepo.GetById(ctx, commentId)
	if err != nil {
		return err
	}

	// atomic: insert like + increment count
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [likecomment: %v]", txErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}
	defer tx.Rollback(ctx)

	txLikeRepo := s.commentLikeRepo.WithTx(tx)

	if insertErr := txLikeRepo.InsertLike(ctx, commentId, userId); insertErr != nil {
		return insertErr
	}

	if incErr := txLikeRepo.IncrementLikeCount(ctx, commentId); incErr != nil {
		return incErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [likecomment: %v]", commitErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return nil
}

func (s *VODCommentService) UnlikeComment(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	// verify comment exists
	_, err := s.commentRepo.GetById(ctx, commentId)
	if err != nil {
		return err
	}

	// atomic: delete like + decrement count
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [unlikecomment: %v]", txErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}
	defer tx.Rollback(ctx)

	txLikeRepo := s.commentLikeRepo.WithTx(tx)

	if delErr := txLikeRepo.DeleteLike(ctx, commentId, userId); delErr != nil {
		return delErr
	}

	if decErr := txLikeRepo.DecrementLikeCount(ctx, commentId); decErr != nil {
		return decErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [unlikecomment: %v]", commitErr)
		return response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return nil
}
