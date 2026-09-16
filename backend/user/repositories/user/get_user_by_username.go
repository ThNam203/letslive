package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) GetByUsername(ctx context.Context, username string) (*domains.User, error) {
	rows, err := r.dbConn.Query(ctx, "select * from users where username = $1", username)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrUserNotFound
		}

		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
