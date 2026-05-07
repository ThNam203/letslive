package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *postgresUserRepo) Create(ctx context.Context, username string, email string, provider domains.AuthProvider) (*domains.User, *response.Response[any]) {
	params := pgx.NamedArgs{
		"username":      username,
		"email":         email,
		"auth_provider": provider,
	}

	row, err := r.dbConn.Query(ctx, "insert into users (username, email, auth_provider) values (@username, @email, @auth_provider) returning *", params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USERNAME_TAKEN, nil, nil, nil,
			)
		}
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	createdUser, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USERNAME_TAKEN, nil, nil, nil,
			)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &createdUser, nil
}
