package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) GetPopular(ctx context.Context, page int, limit int) ([]domains.VOD, int, error) {
	countQuery := `select count(*) from vods where visibility = 'public' and status = 'ready'`
	var total int
	if err := r.dbConn.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		logger.Errorf(ctx, "db count error [getpopularvods: %v]", err)
		return nil, 0, domains.ErrDatabaseQuery
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
		return nil, 0, domains.ErrDatabaseQuery
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getpopularvods: %v]", err)
		return nil, 0, domains.ErrDatabaseIssue
	}

	return vods, total, nil
}
