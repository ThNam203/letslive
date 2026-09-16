package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *postgresUserRepo) Create(ctx context.Context, username string, email string, provider domains.AuthProvider) (*domains.User, error) {
	params := pgx.NamedArgs{
		"username":      username,
		"email":         email,
		"auth_provider": provider,
	}

	row, err := r.dbConn.Query(ctx, "insert into users (username, email, auth_provider) values (@username, @email, @auth_provider) returning *", params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domains.ErrUsernameTaken
		}
		return nil, domains.ErrDatabaseQuery
	}

	createdUser, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domains.ErrUsernameTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrUserNotFound
		}

		return nil, domains.ErrDatabaseIssue
	}

	return &createdUser, nil
}
