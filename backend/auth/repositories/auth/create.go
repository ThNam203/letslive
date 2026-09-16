package auth

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresAuthRepo) Create(ctx context.Context, newAuth domains.Auth) (*domains.Auth, error) {
	params := pgx.NamedArgs{
		"email":         newAuth.Email,
		"password_hash": newAuth.PasswordHash,
		"user_id":       newAuth.UserId,
	}

	rows, err := r.dbConn.Query(ctx, `
		INSERT INTO auths (
			email,
			password_hash,
			user_id
		) values (
			@email, 
			@password_hash, 
			@user_id
		) RETURNING *
	`, params)
	if err != nil {
		logger.Errorf(ctx, "failed to create auth for user %v: %s", newAuth.UserId, err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrAuthNotFound
		}

		logger.Errorf(ctx, "failed to collect created auth row for user %v: %s", newAuth.UserId, err)
		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
