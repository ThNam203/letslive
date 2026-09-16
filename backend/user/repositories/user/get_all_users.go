package user

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5"
)

func (r postgresUserRepo) GetAll(ctx context.Context, page int) ([]domains.User, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, username, email, status, created_at, phone_number, bio, profile_picture, background_picture
		FROM users
		OFFSET $1 LIMIT $2
	`, page*10, 10)

	if err != nil {
		logger.Errorf(ctx, "failed to get all users: %s", err)
		return nil, domains.ErrDatabaseQuery
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		return nil, domains.ErrDatabaseIssue
	}

	return users, nil
}
