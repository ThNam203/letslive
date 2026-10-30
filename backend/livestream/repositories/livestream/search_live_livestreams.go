package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// The normalized title/description expressions must stay identical to the
// index expressions in migration 0006_add_livestream_search.sql.
func (r *postgresLivestreamRepo) SearchLive(ctx context.Context, query string, streamerIds []uuid.UUID, page int, limit int) ([]domains.Livestream, error) {
	offset := limit * page
	sqlQuery := `
		WITH q AS (
			SELECT
				immutable_unaccent(lower($1::text)) AS term,
				'%' || replace(replace(replace(immutable_unaccent(lower($1::text)), '\', '\\'), '%', '\%'), '_', '\_') || '%' AS pattern
		)
		SELECT l.id, l.user_id, l.title, l.description, l.thumbnail_url, l.visibility, l.view_count, l.started_at, l.ended_at, l.created_at, l.updated_at, l.vod_id
		FROM livestreams l
		CROSS JOIN q
		WHERE l.ended_at IS NULL AND l.visibility = 'public'
			AND (
				immutable_unaccent(lower(coalesce(l.title, ''))) % q.term
				OR q.term <% immutable_unaccent(lower(coalesce(l.title, '')))
				OR immutable_unaccent(lower(coalesce(l.title, ''))) LIKE q.pattern ESCAPE '\'
				OR q.term <% immutable_unaccent(lower(coalesce(l.description, '')))
				OR immutable_unaccent(lower(coalesce(l.description, ''))) LIKE q.pattern ESCAPE '\'
				OR l.user_id = ANY($2::uuid[])
			)
		ORDER BY
			greatest(
				similarity(immutable_unaccent(lower(coalesce(l.title, ''))), q.term),
				word_similarity(q.term, immutable_unaccent(lower(coalesce(l.title, '')))),
				0.5 * word_similarity(q.term, immutable_unaccent(lower(coalesce(l.description, ''))))
			) + (CASE WHEN l.user_id = ANY($2::uuid[]) THEN 0.5 ELSE 0 END) DESC,
			l.view_count DESC NULLS LAST,
			l.started_at DESC NULLS LAST,
			l.id
		OFFSET $3 LIMIT $4
	`

	rows, err := r.dbConn.Query(ctx, sqlQuery, query, streamerIds, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [searchlivelivestreams: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf(ctx, "db scan error [searchlivelivestreams: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}

	return livestreams, nil
}
