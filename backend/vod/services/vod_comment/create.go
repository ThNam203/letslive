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
	_, vodErr := s.vodRepo.GetById(ctx, vodId)
	if vodErr != nil {
		return nil, vodErr
	}

	statuses, statusErr := s.userGateway.GetUsersStatuses(ctx, []uuid.UUID{userId})
	if statusErr != nil {
		logger.Errorf(ctx, "failed to check user status for comment creation: %v", statusErr)
		return nil, domains.ErrForbidden
	}
	if status, ok := statuses[userId.String()]; !ok || status == "disabled" {
		return nil, domains.ErrForbidden
	}

	comment := domains.VODComment{
		VODId:   vodId,
		UserId:  userId,
		Content: data.Content,
	}

	// if replying, verify parent exists and belongs to the same VOD
	if data.ParentId != nil {
		parentUUID, err := uuid.FromString(*data.ParentId)
		if err != nil {
			return nil, domains.ErrInvalidInput
		}

		parentComment, parentErr := s.commentRepo.GetById(ctx, parentUUID)
		if parentErr != nil {
			return nil, parentErr
		}

		if parentComment.VODId != vodId {
			return nil, domains.ErrInvalidInput
		}

		comment.ParentId = &parentUUID
	}

	// if this is a reply, create comment + increment parent reply count atomically
	if comment.ParentId != nil {
		return s.createReplyWithTransaction(ctx, comment)
	}

	return s.commentRepo.Create(ctx, comment)
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
