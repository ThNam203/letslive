package vodreaction

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresVODReactionRepo) GetReaction(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*domains.VODReactionType, error) {
	return r.getReaction(ctx, `SELECT reaction FROM vod_reactions WHERE vod_id = $1 AND user_id = $2`, vodId, userId)
}

func (r *postgresVODReactionRepo) GetReactionForUpdate(ctx context.Context, vodId uuid.UUID, userId uuid.UUID) (*domains.VODReactionType, error) {
	return r.getReaction(ctx, `SELECT reaction FROM vod_reactions WHERE vod_id = $1 AND user_id = $2 FOR UPDATE`, vodId, userId)
}

func (r *postgresVODReactionRepo) getReaction(ctx context.Context, query string, vodId uuid.UUID, userId uuid.UUID) (*domains.VODReactionType, error) {
	var reaction domains.VODReactionType
	err := r.db.QueryRow(ctx, query, vodId, userId).Scan(&reaction)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		logger.Errorf(ctx, "db query error [getvodreaction: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}
	return &reaction, nil
}
