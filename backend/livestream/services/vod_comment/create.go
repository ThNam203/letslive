package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/utils"

	"github.com/gofrs/uuid/v5"
)

func (s *VODCommentService) CreateComment(ctx context.Context, data dto.CreateVODCommentRequestDTO, vodId uuid.UUID, userId uuid.UUID) (*domains.VODComment, *response.Response[any]) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, response.NewResponseWithValidationErrors[any](nil, nil, err)
	}

	// verify VOD exists
	_, vodErr := s.vodRepo.GetById(ctx, vodId)
	if vodErr != nil {
		return nil, vodErr
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
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_INVALID_INPUT,
				nil,
				nil,
				nil,
			)
		}

		parentComment, parentErr := s.commentRepo.GetById(ctx, parentUUID)
		if parentErr != nil {
			return nil, parentErr
		}

		if parentComment.VODId != vodId {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_INVALID_INPUT,
				nil,
				nil,
				nil,
			)
		}

		comment.ParentId = &parentUUID
	}

	// if this is a reply, create comment + increment parent reply count atomically
	if comment.ParentId != nil {
		return s.createReplyWithTransaction(ctx, comment)
	}

	return s.commentRepo.Create(ctx, comment)
}

func (s *VODCommentService) createReplyWithTransaction(ctx context.Context, comment domains.VODComment) (*domains.VODComment, *response.Response[any]) {
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [createcomment: %v]", txErr)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
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
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	return createdComment, nil
}
