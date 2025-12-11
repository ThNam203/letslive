package livestream

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamRepo) Update(ctx context.Context, livestream domains.Livestream) (*domains.Livestream, *response.Response[any]) {
	query := `
		UPDATE livestreams
		SET title = $1, description = $2, thumbnail_url = $3, visibility = $4, ended_at = $5, vod_id = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
	`

	rows, err := r.dbConn.Query(ctx, query,
		livestream.Title,
		livestream.Description,
		livestream.ThumbnailURL,
		livestream.Visibility,
		livestream.EndedAt,
		livestream.VODId,
		livestream.Id,
	)
	if err != nil {
		logger.Errorf(ctx, "db query error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_LIVESTREAM_UPDATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	updatedLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_LIVESTREAM_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &updatedLs, nil
}
