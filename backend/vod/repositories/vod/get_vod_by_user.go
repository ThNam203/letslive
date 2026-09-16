package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, error) {
	offset := limit * page
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, status, original_file_url, created_at, updated_at
        from vods
        where user_id = $1
        order by created_at desc
        offset $2 limit $3
    `
	rows, err := r.dbConn.Query(ctx, query, userId, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodbyuser: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getvodbyuser: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return vods, nil
}
