package vod

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODRepo) UpdateStatus(ctx context.Context, vodId uuid.UUID, status domains.VODStatus, playbackUrl *string, thumbnailUrl *string) error {
	query := `
        update vods
        set status = $1, playback_url = COALESCE($2, playback_url), thumbnail_url = COALESCE($3, thumbnail_url), updated_at = now()
        where id = $4
    `
	result, err := r.dbConn.Exec(ctx, query, status, playbackUrl, thumbnailUrl, vodId)
	if err != nil {
		logger.Errorf(ctx, "db query error [updatevodstatus id=%s: %v]", vodId, err)
		return domains.ErrVODUpdateFailed
	}

	if result.RowsAffected() == 0 {
		return domains.ErrVODNotFound
	}

	return nil
}
