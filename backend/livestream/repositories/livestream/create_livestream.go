package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamRepo) Create(ctx context.Context, newLivestream domains.Livestream) (*domains.Livestream, error) {
	query := `
		INSERT INTO livestreams (user_id, title, description, thumbnail_url, visibility)
        	VALUES ($1, $2, $3, $4, $5)
        	RETURNING id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
	`
	rows, err := r.dbConn.Query(ctx, query,
		newLivestream.UserId,
		newLivestream.Title,
		newLivestream.Description,
		newLivestream.ThumbnailURL,
		newLivestream.Visibility,
	)

	if err != nil {
		logger.Errorf(ctx, "db query error [createlivestream: %v]", err)
		return nil, domains.ErrLivestreamCreateFailed
	}

	createdLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createlivestream: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &createdLs, nil
}
