package vodcomment

import (
	"context"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODCommentRepo) HardDelete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	_, err := r.db.Exec(ctx, `DELETE FROM vod_comments WHERE id = $1`, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [harddeletevodcomment: %v]", err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_COMMENT_DELETE_FAILED,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
