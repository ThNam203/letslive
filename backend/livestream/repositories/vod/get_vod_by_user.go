package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	offset := limit * page
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
        from vods
        where user_id = $1
        order by created_at desc
        offset $2 limit $3
    `
	rows, err := r.dbConn.Query(ctx, query, userId, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodbyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getvodbyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return vods, nil
}
