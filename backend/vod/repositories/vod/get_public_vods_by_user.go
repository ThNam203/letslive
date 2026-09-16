package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresVODRepo) GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page, limit int) ([]domains.VOD, error) {
	offset := limit * page
	rows, err := r.dbConn.Query(ctx, `
		SELECT *
		FROM vods
		WHERE user_id = $1 AND visibility = 'public'
		ORDER BY created_at DESC
		OFFSET $2
		LIMIT $3
	`, userId, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db exec error [getpublicvodsbyuser id=%s: %v]", userId, err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getpublicvodsbyuser: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}

	return vods, nil
}
