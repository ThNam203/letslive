package user

import (
	"context"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) GetDisabledUserIds(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.dbConn.Query(ctx, `SELECT id FROM users WHERE status = $1`, domains.UserStatusDisabled)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return nil, domains.ErrDatabaseIssue
	}

	return ids, nil
}
