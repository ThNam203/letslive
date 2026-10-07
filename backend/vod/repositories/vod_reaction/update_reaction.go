package vodreaction

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresVODReactionRepo) UpdateReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID, reaction domains.VODReactionType) error {
	_, err := r.db.Exec(ctx,
		`UPDATE vod_reactions SET reaction = $3, updated_at = now() WHERE vod_id = $1 AND user_id = $2`,
		vodId, userId, reaction,
	)
	if err != nil {
		logger.Errorf(ctx, "db exec error [updatevodreaction: %v]", err)
		return domains.ErrDatabaseIssue
	}
	return nil
}
