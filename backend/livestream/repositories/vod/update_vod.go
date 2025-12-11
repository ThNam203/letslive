package vod

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) Update(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	query := `
        update vods
        set title = $1, description = $2, thumbnail_url = $3, visibility = $4, duration = $5, playback_url = $6, updated_at = now()
        where id = $7
        returning id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
    `
	rows, err := r.dbConn.Query(ctx, query,
		vod.Title, vod.Description, vod.ThumbnailURL, vod.Visibility,
		vod.Duration, vod.PlaybackURL, vod.Id,
	)
	if err != nil {
		logger.Errorf(ctx, "db query error [updatevod id=%s: %v]", vod.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_UPDATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	updatedVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_VOD_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [updatevod id=%s: %v]", vod.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &updatedVod, nil
}
