package user

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/jackc/pgx/v5"
)

func (r postgresUserRepo) GetAll(ctx context.Context, page int) ([]domains.User, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, username, email, status, created_at, display_name, phone_number, bio, profile_picture, background_picture
		FROM users
		OFFSET $1 LIMIT $2
	`, page*10, 10)

	if err != nil {
		logger.Errorf(ctx, "failed to get all users: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return users, nil
}
