package livestream

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.Livestream, error) {
	query := `
		SELECT id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
		FROM livestreams
		WHERE id = $1
	`

	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getlivestreambyid: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	livestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrLivestreamNotFound
		}
		logger.Errorf(ctx, "db scan error [getlivestreambyid: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &livestream, nil
}
