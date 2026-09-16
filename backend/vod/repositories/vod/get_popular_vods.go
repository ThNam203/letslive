package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) GetPopular(ctx context.Context, page int, limit int) ([]domains.VOD, int, *response.Response[any]) {
	countQuery := `select count(*) from vods where visibility = 'public' and status = 'ready'`
	var total int
	if err := r.dbConn.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		logger.Errorf(ctx, "db count error [getpopularvods: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	offset := limit * page
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, status, original_file_url, created_at, updated_at
        from vods
        where visibility = 'public' and status = 'ready'
        order by view_count desc
        offset $1 limit $2
    `

	rows, err := r.dbConn.Query(ctx, query, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getpopularvods: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getpopularvods: %v]", err)
		return nil, 0, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return vods, total, nil
}
