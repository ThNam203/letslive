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

// UpdateComment replaces the content of a comment the caller wrote. The
// superseded content is kept in the comment's edit history, so no version of
// a comment is lost by editing it.
func (s *VODCommentService) UpdateComment(ctx context.Context, data dto.UpdateVODCommentRequestDTO, commentId uuid.UUID, userId uuid.UUID) (*domains.VODComment, error) {
	if err := utils.Validator.Struct(&data); err != nil {
		return nil, fmt.Errorf("%w: %w", domains.ErrInvalidInput, err)
	}

	comment, err := s.commentRepo.GetById(ctx, commentId)
	if err != nil {
		return nil, err
	}

	// a deleted comment has no content to edit, and saying so would confirm
	// it existed
	if comment.IsDeleted {
		return nil, domains.ErrCommentNotFound
	}

	// only the author can edit their own comment
	if comment.UserId != userId {
		return nil, domains.ErrForbidden
	}

	// resubmitting the same text is not an edit: skip the history entry and
	// leave is_edited as it was
	if comment.Content == data.Content {
		return comment, nil
	}

	return s.updateWithTransaction(ctx, *comment, data.Content)
}

func (s *VODCommentService) updateWithTransaction(ctx context.Context, comment domains.VODComment, newContent string) (*domains.VODComment, error) {
	tx, txErr := s.dbPool.Begin(ctx)
	if txErr != nil {
		logger.Errorf(ctx, "failed to begin tx [updatecomment: %v]", txErr)
		return nil, domains.ErrDatabaseIssue
	}
	defer tx.Rollback(ctx)

	// history first: the row records what the comment held before this edit
	if histErr := s.commentEditRepo.WithTx(tx).Insert(ctx, comment.Id, comment.Content); histErr != nil {
		return nil, histErr
	}

	updatedComment, updateErr := s.commentRepo.WithTx(tx).UpdateContent(ctx, comment.Id, newContent)
	if updateErr != nil {
		return nil, updateErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Errorf(ctx, "failed to commit tx [updatecomment: %v]", commitErr)
		return nil, domains.ErrDatabaseIssue
	}

	return updatedComment, nil
}
