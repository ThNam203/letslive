package vod

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresVODRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.VOD, *response.Response[any]) {
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
        from vods
        where id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	vod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_VOD_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getvodbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &vod, nil
}
