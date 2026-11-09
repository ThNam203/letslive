package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

// The normalized title/description expressions must stay identical to the
// index expressions in migration 0004_add_vod_search.sql.
func (r *postgresVODRepo) SearchPublic(ctx context.Context, query string, streamerIds []uuid.UUID, page int, limit int) ([]domains.VOD, error) {
	offset := limit * page
	sqlQuery := `
        with q as (
            select
                immutable_unaccent(lower($1::text)) as term,
                '%' || replace(replace(replace(immutable_unaccent(lower($1::text)), '\', '\\'), '%', '\%'), '_', '\_') || '%' as pattern
        )
        select v.id, v.livestream_id, v.user_id, v.title, v.description, v.thumbnail_url, v.visibility, v.view_count, v.like_count, v.duration, v.playback_url, v.status, v.original_file_url, v.created_at, v.updated_at
        from vods v
        cross join q
        where v.visibility = 'public' and v.status = 'ready'
            and (
                immutable_unaccent(lower(coalesce(v.title, ''))) % q.term
                or q.term <% immutable_unaccent(lower(coalesce(v.title, '')))
                or immutable_unaccent(lower(coalesce(v.title, ''))) like q.pattern escape '\'
                or q.term <% immutable_unaccent(lower(coalesce(v.description, '')))
                or immutable_unaccent(lower(coalesce(v.description, ''))) like q.pattern escape '\'
                or v.user_id = any($2::uuid[])
            )
        order by
            greatest(
                similarity(immutable_unaccent(lower(coalesce(v.title, ''))), q.term),
                word_similarity(q.term, immutable_unaccent(lower(coalesce(v.title, '')))),
                0.5 * word_similarity(q.term, immutable_unaccent(lower(coalesce(v.description, ''))))
            ) + (case when v.user_id = any($2::uuid[]) then 0.5 else 0 end) desc,
            v.view_count desc,
            v.created_at desc,
            v.id
        offset $3 limit $4
    `

	rows, err := r.dbConn.Query(ctx, sqlQuery, query, streamerIds, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [searchpublicvods: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [searchpublicvods: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}

	return vods, nil
}
