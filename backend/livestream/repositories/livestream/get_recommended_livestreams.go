package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

// TODO: implement a recommendation system
func (r *postgresLivestreamRepo) GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]domains.Livestream, error) {
	offset := limit * page
	query := `
		SELECT id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
		FROM livestreams
        	WHERE ended_at IS NULL AND visibility = 'public'
        		AND NOT EXISTS (SELECT 1 FROM disabled_authors da WHERE da.user_id = livestreams.user_id)
        	ORDER BY started_at DESC
        	OFFSET $1 LIMIT $2
	`
	rows, err := r.dbConn.Query(ctx, query, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getalllivestreamings: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getalllivestreamings: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}

	return livestreams, nil
}
