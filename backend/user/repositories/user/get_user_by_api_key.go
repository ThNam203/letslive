package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*domains.User, error) {
	var user domains.User
	rows, err := r.dbConn.Query(ctx, `
		SELECT u.id, u.username, u.email, u.status, u.created_at, u.stream_api_key, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url
		FROM users u
		JOIN livestream_information l ON u.id = l.user_id
		WHERE u.stream_api_key = $1
	`, apiKey)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrUserNotFound
		}

		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
