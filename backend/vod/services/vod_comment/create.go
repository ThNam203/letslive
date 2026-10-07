package vodcomment

import (
	"context"
	"fmt"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/vod/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) CreateComment(ctx context.Context, data dto.CreateVODCommentRequestDTO, vodId uuid.UUID, userId uuid.UUID) (*domains.VODComment, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, fmt.Errorf("%w: %w", domains.ErrInvalidInput, err)
	}

	// verify VOD exists
	vod, vodErr := s.vodRepo.GetById(ctx, vodId)
	if vodErr != nil {
		return nil, vodErr
	}

	comment := domains.VODComment{
		VODId:   vodId,
		UserId:  userId,
		Content: data.Content,
	}

	// if replying, verify parent exists and belongs to the same VOD
	var parentComment *domains.VODComment
	if data.ParentId != nil {
		parentUUID, err := uuid.FromString(*data.ParentId)
		if err != nil {
			return nil, domains.ErrInvalidInput
		}

		parent, parentErr := s.commentRepo.GetById(ctx, parentUUID)
		if parentErr != nil {
			return nil, parentErr
		}
		parentComment = parent

		if parentComment.VODId != vodId {
			return nil, domains.ErrInvalidInput
		}

		comment.ParentId = &parentUUID
	}

	// if this is a reply, create comment + increment parent reply count atomically
	var createdComment *domains.VODComment
	var createErr error
	if comment.ParentId != nil {
		createdComment, createErr = s.createReplyWithTransaction(ctx, comment)
	} else {
		createdComment, createErr = s.commentRepo.Create(ctx, comment)
	}
	if createErr != nil {
		return nil, createErr
	}

	// the request context is cancelled once the handler returns, so detach it
	notifyCtx := context.WithoutCancel(ctx)
	go s.notifyNewComment(notifyCtx, *vod, parentComment, *createdComment)

	return createdComment, nil
}

func (s *VODCommentService) createReplyWithTransaction(ctx context.Context, comment domains.VODComment) (*domains.VODComment, error) {
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [createcomment: %v]", txErr)
		return nil, domains.ErrDatabaseIssue
	}
	defer tx.Rollback(ctx)

	txCommentRepo := s.commentRepo.WithTx(tx)

	createdComment, createErr := txCommentRepo.Create(ctx, comment)
	if createErr != nil {
		return nil, createErr
	}

	if incErr := txCommentRepo.IncrementReplyCount(ctx, *comment.ParentId); incErr != nil {
		return nil, incErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [createcomment: %v]", commitErr)
		return nil, domains.ErrDatabaseIssue
	}

	return createdComment, nil
}
