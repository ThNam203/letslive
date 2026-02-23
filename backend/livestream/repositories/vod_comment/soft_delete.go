package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) SoftDelete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	_, err := r.db.Exec(ctx, `
		UPDATE vod_comments SET is_deleted = true, updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [softdeletevodcomment: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_COMMENT_DELETE_FAILED,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
