package vod

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	query := `
        insert into vods (livestream_id, user_id, title, description, thumbnail_url, visibility, duration, playback_url, view_count, created_at)
        values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        returning id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
    `
	rows, err := r.dbConn.Query(ctx, query,
		vod.LivestreamId, vod.UserId, vod.Title, vod.Description, vod.ThumbnailURL,
		vod.Visibility, vod.Duration, vod.PlaybackURL, vod.ViewCount, vod.CreatedAt,
	)

	if err != nil {
		// todo: check for specific db errors like fk violations if possible
		logger.Errorf(ctx, "db query error [createvod: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_CREATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	createdVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createvod: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &createdVod, nil
}
