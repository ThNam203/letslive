package livestream

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamRepo) Update(ctx context.Context, livestream domains.Livestream) (*domains.Livestream, error) {
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
		return nil, domains.ErrLivestreamUpdateFailed
	}

	updatedLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrLivestreamNotFound
		}
		logger.Errorf(ctx, "db scan error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, domains.ErrDatabaseIssue
	}
	return &updatedLs, nil
}
