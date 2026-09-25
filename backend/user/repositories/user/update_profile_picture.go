package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) UpdateProfilePicture(ctx context.Context, userId uuid.UUID, path string) error {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET profile_picture = $1 WHERE id = $2", path, userId)
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
