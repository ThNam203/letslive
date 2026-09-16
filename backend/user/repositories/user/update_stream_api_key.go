package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) error {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET stream_api_key = $1 WHERE id = $2", newKey, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.ErrUserNotFound
		}

		return domains.ErrDatabaseQuery
	} else if result.RowsAffected() == 0 {
		return domains.ErrUserNotFound
	}

	return nil
}
