package vod

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresVODRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.VOD, error) {
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, status, original_file_url, created_at, updated_at
        from vods
        where id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodbyid: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	vod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrVODNotFound
		}
		logger.Errorf(ctx, "db scan error [getvodbyid: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &vod, nil
}
