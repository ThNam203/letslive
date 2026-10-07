package vodreaction

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODReactionRepo) InsertReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID, reaction domains.VODReactionType) (bool, error) {
	result, err := r.db.Exec(ctx,
		`INSERT INTO vod_reactions (vod_id, user_id, reaction) VALUES ($1, $2, $3) ON CONFLICT (vod_id, user_id) DO NOTHING`,
		vodId, userId, reaction,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [insertvodreaction: %v]", err)
		return false, domains.ErrDatabaseIssue
	}
	return result.RowsAffected() == 1, nil
}
