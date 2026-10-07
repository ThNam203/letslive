package vodreaction

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODReactionRepo) DeleteReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*domains.VODReactionType, error) {
	var removed domains.VODReactionType
	err := r.db.QueryRow(ctx,
		`DELETE FROM vod_reactions WHERE vod_id = $1 AND user_id = $2 RETURNING reaction`,
		vodId, userId,
	).Scan(&removed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		logger.Errorf(ctx, "db exec error [deletevodreaction: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &removed, nil
}
