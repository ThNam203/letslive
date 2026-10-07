package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODRepo) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, error) {
	// the upload flow picks the id up front so the raw file can be stored under
	// it, so honour a caller-supplied one; the livestream flow leaves it unset
	// and lets the database generate it
	var id *uuid.UUID
	if !vod.Id.IsNil() {
		id = &vod.Id
	}

	query := `
        insert into vods (id, livestream_id, user_id, title, description, thumbnail_url, visibility, duration, playback_url, view_count, status, original_file_url, created_at)
        values (coalesce($1::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        returning id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, like_count, duration, playback_url, status, original_file_url, created_at, updated_at
    `
	rows, err := r.dbConn.Query(ctx, query,
		id, vod.LivestreamId, vod.UserId, vod.Title, vod.Description, vod.ThumbnailURL,
		vod.Visibility, vod.Duration, vod.PlaybackURL, vod.ViewCount, vod.Status, vod.OriginalFileURL, vod.CreatedAt,
	)

	if err != nil {
		// todo: check for specific db errors like fk violations if possible
		logger.Errorf(ctx, "db query error [createvod: %v]", err)
		return nil, domains.ErrVODCreateFailed
	}

	createdVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createvod: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &createdVod, nil
}
